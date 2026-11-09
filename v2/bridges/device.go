package bridges

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/jimjibone/log"
	"github.com/jimjibone/wh/v2/bridges/services"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
)

// RootEndpointID is the endpoint holding device-wide services: info, online,
// update, identify and the like.
const RootEndpointID = "0"

// Device is a device provided by a bridge. It has endpoints, one per physical
// or logical unit (e.g. each output of a two-channel relay), each holding
// services. The root endpoint "0" always exists and holds the Info and Online
// services.
type Device struct {
	log *log.Context
	id  string
	typ clientsapi.Device_DeviceType
	// sendMu guards sendState. It is separate from mu because the send loop
	// must never wait for mu: SendFullState holds mu while reading attributes,
	// whose updates queue to the send loop.
	sendMu    sync.Mutex
	sendState func(state *clientsapi.Device)

	mu        sync.RWMutex // guards endpoints
	endpoints []*Endpoint  // in the order created

	updates chan update
	wg      sync.WaitGroup
	close   func()
	done    <-chan struct{} // closed by Close

	// The device's info service. Note that the default info service will be
	// overwritten if a new info service is added to the root endpoint.
	Info *services.Info
	// The device's online service. Note that the default online service will be
	// overwritten if a new online service is added to the root endpoint.
	Online *services.Online
}

// update is one item for the device's send loop: a full state, a service
// update, an endpoint label change or an event.
type update struct {
	full       *clientsapi.Device
	endpointID string
	service    *clientsapi.Service
	label      *string
	event      *clientsapi.Event
}

// Endpoint is one unit of a device and holds its services. Create endpoints
// with Device.Endpoint.
type Endpoint struct {
	dev      *Device
	id       string
	label    string
	services []services.Service
}

// NewDevice creates a device with the specified ID and device type. The ID
// must follow the device ID convention, "<protocol><unique-id>" in lowercase
// [a-z0-9] (see DeviceID), and be stable for the life of the device. After
// creation, the device should be cleaned up via the Close method. Use
// AddService (root endpoint) or Endpoint(id).AddService to add services.
func NewDevice(id string, typ clientsapi.Device_DeviceType) *Device {
	ctx, close := context.WithCancel(context.Background())
	dev := &Device{
		log:     log.NewContext(log.DefaultLogger, id, log.DebugLevel),
		id:      id,
		typ:     typ,
		updates: make(chan update, 1),
		close:   close,
		done:    ctx.Done(),
		Info:    services.NewInfo(""),
		Online:  services.NewOnline(""),
	}
	dev.wg.Add(1)
	go dev.run(ctx)

	// Set default name and online state for devices that don't explicitly
	// implement them.
	dev.AddService(dev.Info, dev.Online)
	dev.Info.Name.Set(id)
	dev.Online.Online.Set(true)
	dev.Online.LastSeen.Set(time.Now().Truncate(time.Second))

	return dev
}

// Close stops the device. Updates made after Close are dropped.
func (dev *Device) Close() {
	dev.close()
	dev.wg.Wait()
}

// ID returns the device's ID.
func (dev *Device) ID() string { return dev.id }

// Root returns the root endpoint "0".
func (dev *Device) Root() *Endpoint { return dev.Endpoint(RootEndpointID) }

// Endpoint returns the endpoint with the ID, creating it if needed. IDs must
// be unique within the device and stable across restarts, e.g. "1", "2" or a
// native unit ID such as "l1".
func (dev *Device) Endpoint(id string) *Endpoint {
	dev.mu.Lock()
	defer dev.mu.Unlock()
	for _, ep := range dev.endpoints {
		if ep.id == id {
			return ep
		}
	}
	ep := &Endpoint{dev: dev, id: id}
	dev.endpoints = append(dev.endpoints, ep)
	return ep
}

// Endpoints returns the device's endpoints in the order created.
func (dev *Device) Endpoints() []*Endpoint {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	return slices.Clone(dev.endpoints)
}

// AddService adds services to the root endpoint.
func (dev *Device) AddService(srvs ...services.Service) {
	dev.Root().AddService(srvs...)
}

// ID returns the endpoint's ID.
func (ep *Endpoint) ID() string { return ep.id }

// SetLabel sets the endpoint's label, e.g. "Output 1". Users may override it.
func (ep *Endpoint) SetLabel(label string) {
	ep.dev.mu.Lock()
	changed := ep.label != label
	ep.label = label
	ep.dev.mu.Unlock()
	if changed {
		ep.dev.queue(update{endpointID: ep.id, label: &label})
	}
}

// AddService adds services to the endpoint. Adding a service ID that already
// exists on the endpoint replaces it.
func (ep *Endpoint) AddService(srvs ...services.Service) {
	ep.dev.mu.Lock()
	defer ep.dev.mu.Unlock()
	for _, srv := range srvs {
		if ep.id == RootEndpointID {
			switch s := srv.(type) {
			case *services.Info:
				ep.dev.Info = s
			case *services.Online:
				ep.dev.Online = s
			}
		}
		srv.Push(
			func(s *clientsapi.Service) { ep.dev.queue(update{endpointID: ep.id, service: s}) },
			func(e *clientsapi.Event) {
				e.EndpointId = ep.id
				ep.dev.queue(update{endpointID: ep.id, event: e})
			},
		)
		ep.services = slices.DeleteFunc(ep.services, func(s services.Service) bool { return s.ID() == srv.ID() })
		ep.services = append(ep.services, srv)
	}
}

// Service returns the endpoint's service with the ID, or nil.
func (ep *Endpoint) Service(id string) services.Service {
	ep.dev.mu.RLock()
	defer ep.dev.mu.RUnlock()
	return ep.service(id)
}

func (ep *Endpoint) service(id string) services.Service {
	for _, srv := range ep.services {
		if srv.ID() == id {
			return srv
		}
	}
	return nil
}

// Init is called when the device is added to the client. This should be
// used to keep the sendState func for later use and may be used for other
// initialisation tasks.
// sendState sends a device state message to the server. This may either contain
// all services, or a diff of those changed since the last call of SendState by
// the device. The client will indicate if a full state should be sent by
// calling SendFullState on the device.
func (dev *Device) Init(sendState func(state *clientsapi.Device)) {
	dev.sendMu.Lock()
	defer dev.sendMu.Unlock()
	dev.sendState = sendState
}

// SendFullState is called when a full device state should be sent to the
// server. This is typically done just after the client connects to the
// server or this device is added to the client after connection.
func (dev *Device) SendFullState() {
	dev.mu.RLock()
	pb := &clientsapi.Device{
		Id:        dev.id,
		FullState: true,
		Typ:       dev.typ,
		Endpoints: []*clientsapi.Endpoint{},
	}
	for _, ep := range dev.endpoints {
		pbep := &clientsapi.Endpoint{Id: ep.id, Label: ep.label, Services: []*clientsapi.Service{}}
		for _, srv := range ep.services {
			pbep.Services = append(pbep.Services, srv.Pb())
		}
		pb.Endpoints = append(pb.Endpoints, pbep)
	}
	dev.mu.RUnlock()
	dev.queue(update{full: pb})
}

func (dev *Device) queue(u update) {
	select {
	case dev.updates <- u:
	case <-dev.done:
	}
}

// HandleAction is called in its own goroutine when an ActionRequest is
// received from the server. It routes the request to the service on the
// requested endpoint, which applies the write or runs the command. The
// service may send ActionResponse feedback at any time, which is useful if
// the request may take a long time or time out. When it returns, a final
// ActionResponse is sent: COMPLETE with the command result if the error is
// nil, otherwise an error status with the error message.
func (dev *Device) HandleAction(request *clientsapi.ActionRequest, feedback func(*clientsapi.ActionResponse)) ([]*clientsapi.FieldValue, error) {
	srv, err := dev.find(request.GetEndpointId(), request.GetServiceId())
	if err != nil {
		return nil, err
	}
	return srv.HandleAction(request, feedback)
}

// HandleImage is called in its own goroutine when an ImageRequest is
// received from the server and routes it to the service like HandleAction.
// When it returns, a final ImageResponse is sent: COMPLETE with the returned
// data if the error is nil, otherwise ERROR with the error message.
func (dev *Device) HandleImage(request *clientsapi.ImageRequest, feedback func(*clientsapi.ImageResponse)) ([]byte, error) {
	srv, err := dev.find(request.GetEndpointId(), request.GetServiceId())
	if err != nil {
		return nil, err
	}
	return srv.HandleImage(request, feedback)
}

func (dev *Device) find(endpointID, serviceID string) (services.Service, error) {
	dev.mu.RLock()
	defer dev.mu.RUnlock()
	for _, ep := range dev.endpoints {
		if ep.id != endpointID {
			continue
		}
		if srv := ep.service(serviceID); srv != nil {
			return srv, nil
		}
		return nil, &services.InvalidError{Err: fmt.Errorf("service %q not found on endpoint %q", serviceID, endpointID)}
	}
	return nil, &services.InvalidError{Err: fmt.Errorf("endpoint %q not found", endpointID)}
}

// run batches partial updates and events and sends them at most every
// tickInterval, or at once when ten have queued. A full state replaces the
// pending partial updates; pending events are sent first, never dropped.
func (dev *Device) run(ctx context.Context) {
	defer dev.wg.Done()

	tickInterval := 50 * time.Millisecond
	ticker := time.NewTicker(tickInterval)
	ticker.Stop() // stop now as updates will restart it.
	defer ticker.Stop()

	var (
		count  int
		order  []string                                  // endpoint IDs in the order first updated
		cache  map[string]*clientsapi.Endpoint           // key=endpoint ID
		srvs   map[string]map[string]*clientsapi.Service // key=endpoint ID, service ID
		events []*clientsapi.Event
	)
	resetCache := func() {
		count = 0
		order = nil
		cache = make(map[string]*clientsapi.Endpoint)
		srvs = make(map[string]map[string]*clientsapi.Service)
	}
	resetCache()

	send := func(what string, pb *clientsapi.Device) {
		dev.sendMu.Lock()
		sendState := dev.sendState
		dev.sendMu.Unlock()
		if sendState == nil {
			dev.log.Debugf("warning not sending %s (device not registered with client)\n%s", what, PrettyDevice("  ", pb))
			return
		}
		dev.log.Debugf("sending %s\n%s", what, PrettyDevice("  ", pb))
		sendState(pb)
	}
	sendCache := func() {
		if len(cache) == 0 && len(events) == 0 {
			return
		}
		pb := &clientsapi.Device{Id: dev.id, Typ: dev.typ, Events: events}
		for _, id := range order {
			pb.Endpoints = append(pb.Endpoints, cache[id])
		}
		send("update", pb)
		events = nil
		resetCache()
	}
	endpoint := func(id string) *clientsapi.Endpoint {
		ep, found := cache[id]
		if !found {
			ep = &clientsapi.Endpoint{Id: id}
			cache[id] = ep
			srvs[id] = make(map[string]*clientsapi.Service)
			order = append(order, id)
		}
		return ep
	}

	for {
		select {
		case <-ctx.Done():
			return

		case u := <-dev.updates:
			switch {
			case u.full != nil:
				if len(events) > 0 {
					send("events", &clientsapi.Device{Id: dev.id, Typ: dev.typ, Events: events})
					events = nil
				}
				resetCache()
				ticker.Stop()
				send("full state", u.full)
				continue
			case u.event != nil:
				events = append(events, u.event)
			case u.label != nil:
				endpoint(u.endpointID).Label = *u.label
			case u.service != nil:
				ep := endpoint(u.endpointID)
				if srv, found := srvs[u.endpointID][u.service.GetId()]; found {
					mergeService(srv, u.service)
				} else {
					srvs[u.endpointID][u.service.GetId()] = u.service
					ep.Services = append(ep.Services, u.service)
				}
			}

			// Send soon, or now if there are 10 updates.
			count++
			if count < 10 {
				ticker.Reset(tickInterval)
			} else {
				sendCache()
				ticker.Stop()
			}

		case <-ticker.C:
			sendCache()
			ticker.Stop()
		}
	}
}

// mergeService merges a partial service update into a pending one: later
// attribute updates replace earlier ones of the same ID.
func mergeService(srv, update *clientsapi.Service) {
	srv.Typ = update.Typ
	srv.Alias = update.Alias
	for _, up := range update.Attrs {
		i := slices.IndexFunc(srv.Attrs, func(a *clientsapi.Attribute) bool { return a.GetId() == up.GetId() })
		switch {
		case i < 0:
			srv.Attrs = append(srv.Attrs, up)
		case up.Spec == nil && srv.Attrs[i].Spec != nil:
			// Keep the pending spec and perms; take the new value and time.
			srv.Attrs[i].Value = up.Value
			srv.Attrs[i].Time = up.Time
		default:
			srv.Attrs[i] = up
		}
	}
}
