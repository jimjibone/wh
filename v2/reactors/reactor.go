package reactors

//go:generate go run ../internal/whgen -kind reactor -out zz_views.go

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v2/bridges"
	"github.com/jimjibone/wh/v2/clients"
	"github.com/jimjibone/wh/v2/waiter"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateCallback is called for each non-nil device update received from the
// server, before the update is dispatched to any views. It is invoked
// outside of the reactor's lock so the implementer is free to create views in
// response to the update.
type UpdateCallback func(update *clientsapi.Device)

// Reactor follows device state and events from the server and acts on them
// through typed service views (rc.Light, rc.Button, ...). It is also a bridge,
// so it can provide devices of its own.
type Reactor struct {
	log    *log.Context
	bridge *bridges.Bridge
	ready  *waiter.Waiter // ready when we're connected to the server and have received the first batch of device states

	mu           sync.RWMutex
	devices      map[string]*deviceState // key=device ID
	views        map[serviceKey][]*ServiceView
	onUpdate     UpdateCallback
	onEventsLost func()
	cursor       *clientsapi.EventCursor // last event seen, resumed from on reconnect
	now          func() time.Time

	serviceMu sync.RWMutex
	service   clientsapi.ClientServiceClient
}

func NewReactor(config bridges.BridgeConfig, opts ...clients.ClientOption) *Reactor {
	rc := &Reactor{
		log:     log.NewContext(log.DefaultLogger, "reactor", log.DebugLevel),
		ready:   waiter.NewWaiter(),
		devices: make(map[string]*deviceState),
		views:   make(map[serviceKey][]*ServiceView),
		now:     time.Now,
	}
	opts = append(opts, clients.WithConnectionHandler(rc.runloop))
	rc.bridge = bridges.NewBridge(config, opts...)
	return rc
}

// Returns a channel which is closed when the client connects to the server
// and has received the initial device states.
func (rc *Reactor) Ready() <-chan struct{} {
	return rc.ready.Wait()
}

func (rc *Reactor) Bridge() *bridges.Bridge {
	return rc.bridge
}

func (rc *Reactor) Client() *clients.Client {
	return rc.bridge.Client()
}

// OnUpdate registers a callback which is invoked for each non-nil device update
// received from the server, before the update is dispatched to any views.
// Only one callback may be registered; calling OnUpdate again replaces it.
func (rc *Reactor) OnUpdate(callback UpdateCallback) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.onUpdate = callback
}

// OnEventsLost registers a callback for when the server reports that events
// were lost while the reactor was disconnected (the server restarted, or the
// gap was longer than its buffer).
func (rc *Reactor) OnEventsLost(callback func()) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.onEventsLost = callback
}

// Generic returns an untyped view of any service, read and driven by its raw
// attributes, commands and events. Use it for generic services, or for
// attributes a typed view does not cover.
func (rc *Reactor) Generic(deviceID, endpointID, serviceID string) *ServiceView {
	return rc.view(deviceID, endpointID, []string{serviceID}, "", clientsapi.Service_UNDEFINED)
}

// view registers a view. serviceID optionally overrides defaultID.
func (rc *Reactor) view(deviceID, endpointID string, serviceID []string, defaultID string, typ clientsapi.Service_ServiceType) *ServiceView {
	if deviceID == "" {
		panic("device id must be defined")
	}
	if endpointID == "" {
		panic("endpoint id must be defined")
	}
	if len(serviceID) > 1 {
		panic("only one serviceID argument is allowed")
	}
	id := defaultID
	if len(serviceID) == 1 {
		id = serviceID[0]
	}
	if id == "" {
		panic("service id must be defined")
	}
	v := &ServiceView{rc: rc, key: serviceKey{deviceID, endpointID, id}, typ: typ}
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.views[v.key] = append(rc.views[v.key], v)
	return v
}

func (rc *Reactor) Run() error {
	return rc.bridge.Run()
}

func (rc *Reactor) requestAction(ctx context.Context, req *clientsapi.ActionRequest, handler func(resp *clientsapi.ActionResponse)) error {
	rc.serviceMu.RLock()
	defer rc.serviceMu.RUnlock()

	if rc.service == nil {
		return clients.ErrNotConnected
	}

	stream, err := rc.service.SendAction(ctx, req)
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		handler(resp)
	}
}

func (rc *Reactor) runloop(ctx context.Context, conn *grpc.ClientConn) {
	rc.log.Infof("stream started")
	defer rc.log.Infof("stream finished")

	// Create a new shared service for action requests.
	rc.serviceMu.Lock()
	rc.service = clientsapi.NewClientServiceClient(conn)
	rc.serviceMu.Unlock()

	service := clientsapi.NewClientServiceClient(conn)
	stream, err := service.DeviceStream(ctx)
	if err != nil {
		rc.log.Errorf("failed to start reactor stream: %s", err)
		return
	}

	// Start the stream. After a reconnect, resume events from the last one
	// seen so ones raised meanwhile are replayed.
	rc.mu.RLock()
	req := &clientsapi.DeviceStreamRequest{EventsSince: rc.cursor}
	rc.mu.RUnlock()
	if err := stream.Send(req); err != nil {
		rc.log.Errorf("failed to start reactor stream: %s", err)
		return
	}

	// Reset the waiter on exit so clients can wait for reconnection.
	defer rc.ready.Reset()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		res, err := stream.Recv()
		if err != nil {
			code := status.Code(err)
			if code == codes.Unavailable || code == codes.Canceled {
				rc.log.Debugf("stream closed: %s", err)
			} else {
				rc.log.Errorf("failed to recv reactor request: %s", err)
			}
			return
		}
		rc.handle(res)
	}
}

// handle applies one stream response: state to the mirror, then update and
// event handlers, which run without the reactor's lock held.
func (rc *Reactor) handle(res *clientsapi.DeviceStreamResponse) {
	if res.GetEventsLost() {
		rc.log.Warnf("server reports events lost while disconnected")
		rc.mu.RLock()
		lost := rc.onEventsLost
		rc.mu.RUnlock()
		if lost != nil {
			lost()
		}
	}

	if id := res.GetDeviceRemoved(); id != "" {
		rc.mu.Lock()
		old := rc.devices[id]
		delete(rc.devices, id)
		var calls []func()
		for _, ep := range old.pbOrNil().GetEndpoints() {
			for _, srv := range ep.GetServices() {
				calls = append(calls, rc.updateCalls(serviceKey{id, ep.GetId(), srv.GetId()}, true)...)
			}
		}
		rc.mu.Unlock()
		for _, call := range calls {
			call()
		}
		return
	}

	dev := res.GetDevice()
	if dev == nil {
		// Detect when the full state has finished being streamed (the first
		// empty message is our cue).
		rc.ready.Done()
		return
	}

	rc.mu.RLock()
	onUpdate := rc.onUpdate
	rc.mu.RUnlock()
	if onUpdate != nil {
		onUpdate(dev)
	}

	rc.mu.Lock()
	var calls []func()
	if len(dev.GetEndpoints()) > 0 || dev.GetFullState() {
		st := rc.devices[dev.GetId()]
		if st == nil {
			st = &deviceState{}
			rc.devices[dev.GetId()] = st
		}
		for _, t := range st.apply(dev) {
			calls = append(calls, rc.updateCalls(t.key, t.changed)...)
		}
	}
	now := rc.now()
	for _, e := range dev.GetEvents() {
		if e.GetEpoch() != "" {
			rc.cursor = &clientsapi.EventCursor{Epoch: e.GetEpoch(), Seq: e.GetSeq()}
		}
		calls = append(calls, rc.eventCalls(dev.GetId(), e, now)...)
	}
	rc.mu.Unlock()

	for _, call := range calls {
		call()
	}
}

// updateCalls returns the update handlers for a service. The caller holds
// rc.mu.
func (rc *Reactor) updateCalls(key serviceKey, changed bool) []func() {
	var calls []func()
	for _, v := range rc.views[key] {
		if h := v.onUpdate; h != nil {
			calls = append(calls, func() { h(changed) })
		}
	}
	return calls
}

// eventCalls returns the event handlers that take an event. The caller holds
// rc.mu.
func (rc *Reactor) eventCalls(deviceID string, e *clientsapi.Event, now time.Time) []func() {
	key := serviceKey{deviceID, e.GetEndpointId(), e.GetServiceId()}
	info := eventInfo(e)
	var calls []func()
	for _, v := range rc.views[key] {
		for _, h := range v.events {
			if h.wants(e, info, now) {
				fn := h.fn
				calls = append(calls, func() { fn(e, info) })
			}
		}
	}
	return calls
}
