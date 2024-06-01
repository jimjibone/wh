package devices

import (
	"context"
	"fmt"
	"sync"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
	"github.com/jimjibone/woodhouse-4/log"
	"github.com/jimjibone/woodhouse-4/wh/v1/devices/services"
)

// type Device interface {
// 	// ID returns the device's ID.
// 	ID() string

// 	// Init is called when the device is added to the client. This should be
// 	// used to keep the sendState func for later use and may be used for other
// 	// initialisation tasks.
// 	// sendState sends a device state message to the server. This may either contain
// 	// all services, or a diff of those changed since the last call of SendState by
// 	// the device. The client will indicate if a full state should be sent by
// 	// calling SendFullState on the device.
// 	Init(sendState func(state *clientsapi.Device))

// 	// SendFullState is called when a full device state should be sent to the
// 	// server. This is typically done just after the client connects to the
// 	// server or this device is added to the client after connection.
// 	SendFullState()

// 	// HandleAction is called in its own goroutine when an ActionRequest is
// 	// received from the server. During this the implementer should forward the
// 	// request to the contained services and attributes. The implementer is
// 	// responsible for ensuring safety of concurrency. During this time the
// 	// implementer may send ActionResponse feedback at any time using the
// 	// feedback func, which is useful if the request has the potential to take a
// 	// long time or timeout, for example. When the function returns a final
// 	// ActionResponse will be sent back to the server. If the returned error is
// 	// nil the response status will be COMPLETE, otherwise ERR with the details
// 	// field containing the error message.
// 	HandleAction(request *clientsapi.ActionRequest, feedback func(*clientsapi.ActionResponse)) error
// }

type Device struct {
	log            *log.Context
	id             string
	typ            clientsapi.Device_DeviceType
	sendState      func(state *clientsapi.Device)
	services       map[string]services.Service
	deviceUpdates  chan *clientsapi.Device
	serviceUpdates chan *clientsapi.Service
	wg             sync.WaitGroup
	close          func()
}

func NewDevice(id string, typ clientsapi.Device_DeviceType) *Device {
	ctx, close := context.WithCancel(context.Background())
	dev := &Device{
		log:            log.NewContext(log.DefaultLogger, id, log.DebugLevel),
		id:             id,
		typ:            typ,
		services:       make(map[string]services.Service),
		deviceUpdates:  make(chan *clientsapi.Device, 1),
		serviceUpdates: make(chan *clientsapi.Service, 1),
		close:          close,
	}
	dev.wg.Add(1)
	go dev.run(ctx)
	return dev
}

func (dev *Device) Close() {
	dev.close()
	dev.wg.Wait()
}

// AddService adds the services to the device.
func (dev *Device) AddService(srvs ...services.Service) {
	for _, srv := range srvs {
		srv.Push(dev.pusher)
		dev.services[srv.ID()] = srv
	}
}

// ID returns the device's ID.
func (dev *Device) ID() string { return dev.id }

// Init is called when the device is added to the client. This should be
// used to keep the sendState func for later use and may be used for other
// initialisation tasks.
// sendState sends a device state message to the server. This may either contain
// all services, or a diff of those changed since the last call of SendState by
// the device. The client will indicate if a full state should be sent by
// calling SendFullState on the device.
func (dev *Device) Init(sendState func(state *clientsapi.Device)) { dev.sendState = sendState }

// SendFullState is called when a full device state should be sent to the
// server. This is typically done just after the client connects to the
// server or this device is added to the client after connection.
func (dev *Device) SendFullState() {
	pb := &clientsapi.Device{
		Id:        dev.id,
		FullState: true,
		Typ:       dev.typ,
		Services:  []*clientsapi.Service{},
	}
	for _, srv := range dev.services {
		pb.Services = append(pb.Services, srv.Pb())
	}
	dev.deviceUpdates <- pb
}

func (dev *Device) pusher(srv *clientsapi.Service) {
	dev.serviceUpdates <- srv
}

// HandleAction is called in its own goroutine when an ActionRequest is
// received from the server. During this the implementer should forward the
// request to the contained services and attributes. The implementer is
// responsible for ensuring safety of concurrency. During this time the
// implementer may send ActionResponse feedback at any time using the
// feedback func, which is useful if the request has the potential to take a
// long time or timeout, for example. When the function returns a final
// ActionResponse will be sent back to the server. If the returned error is
// nil the response status will be COMPLETE, otherwise ERR with the details
// field containing the error message.
func (dev *Device) HandleAction(request *clientsapi.ActionRequest, feedback func(*clientsapi.ActionResponse)) error {
	srv, found := dev.services[request.GetServiceId()]
	if !found {
		return fmt.Errorf("service not found")
	}
	return srv.Action(request, feedback)
}

func (dev *Device) run(ctx context.Context) {
	defer dev.wg.Done()

	tickInterval := 50 * time.Millisecond
	ticker := time.NewTicker(tickInterval)
	ticker.Stop() // stop now as updates will restart it.
	defer ticker.Stop()

	count := 0
	cache := make(map[string]*clientsapi.Service)

	resetCache := func() {
		count = 0
		cache = make(map[string]*clientsapi.Service)
	}
	sendCache := func() {
		// Send the cached update if it's not empty.
		if len(cache) > 0 {
			pb := &clientsapi.Device{
				Id:       dev.id,
				Typ:      dev.typ,
				Services: []*clientsapi.Service{},
			}
			for _, srv := range cache {
				pb.Services = append(pb.Services, srv)
			}

			if dev.sendState != nil {
				dev.log.Debugf("sending service update typ:%q, services:%d\n%s", dev.typ, len(cache), services.PrettyServices("  ", pb.Services))
				dev.sendState(pb)
			} else {
				dev.log.Debugf("warning not sending service update (device not registered with client) typ:%q, services:%d\n%s", dev.typ, len(cache), services.PrettyServices("  ", pb.Services))
			}

			// Reset the cache.
			resetCache()
		}
	}

	// This crazy thing enables rate limiting of updates sent to the server.
	for {
		select {
		case <-ctx.Done():
			return

		case state := <-dev.deviceUpdates:
			resetCache()
			if dev.sendState != nil {
				dev.log.Debugf("sending full state typ:%q, services:%d\n%s", dev.typ, len(cache), services.PrettyServices("  ", state.Services))
				dev.sendState(state)
			} else {
				dev.log.Debugf("warning not sending full state (device not registered with client) typ:%q, services:%d\n%s", dev.typ, len(cache), services.PrettyServices("  ", state.Services))
			}

		case update := <-dev.serviceUpdates:
			// Merge the new update into the cache.
			if srv, found := cache[update.GetId()]; found {
				srv.Id = update.Id
				srv.Typ = update.Typ
				srv.Alias = update.Alias
				for _, up := range update.Attrs {
					found := false
					for i, attr := range srv.Attrs {
						if up.GetId() == attr.GetId() {
							found = true
							srv.Attrs[i] = up
							break
						}
					}
					if !found {
						srv.Attrs = append(srv.Attrs, up)
					}
				}
			} else {
				cache[update.GetId()] = update
			}

			// Increment the count and send immediately if there are 10 updates.
			count++
			if count < 10 {
				// Reset the ticker to send soon.
				// log.Infof("reset ticker %s", dev.id)
				ticker.Reset(tickInterval)
			} else {
				// Send immediately!
				// log.Infof("sending due to count - stop ticker %s", dev.id)
				sendCache()
				ticker.Stop()
			}

		case <-ticker.C:
			// The ticker has finally expired. Send the updates!
			// log.Infof("sending due to ticker - stop ticker %s", dev.id)
			sendCache()
			ticker.Stop()
		}
	}
}
