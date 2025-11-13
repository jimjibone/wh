package reactors

import (
	"context"
	"fmt"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type GenericService struct {
	id        string
	onUpdate  func(changed bool)
	requester requester
	wait      *Waiter
	onliner

	bools  map[string]bool
	floats map[string]float64
}

// Initialises the service.
func (srv *GenericService) init(serviceID string, requester requester) {
	srv.id = serviceID
	srv.requester = requester
	srv.wait = NewWaiter()
	srv.bools = make(map[string]bool)
	srv.floats = make(map[string]float64)
}

// Handle the update. Returns true if the values changed.
func (srv *GenericService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	srv.id = update.GetId()
	for _, attr := range update.Attrs {
		if attr.GetBool() != nil {
			if prev, found := srv.bools[attr.GetId()]; found {
				if prev != attr.GetBool().GetValue() {
					changed = true
					srv.bools[attr.GetId()] = attr.GetBool().GetValue()
				}
			} else {
				changed = true
				srv.bools[attr.GetId()] = attr.GetBool().GetValue()
			}
		} else if attr.GetFloat() != nil {
			if prev, found := srv.floats[attr.GetId()]; found {
				if prev != attr.GetFloat().GetValue() {
					changed = true
					srv.floats[attr.GetId()] = attr.GetFloat().GetValue()
				}
			} else {
				changed = true
				srv.floats[attr.GetId()] = attr.GetFloat().GetValue()
			}
		}
	}
	if srv.onUpdate != nil {
		srv.onUpdate(changed)
	}
	srv.wait.Done()
	return changed
}

// Sets a handler to be called when the service is updated.
func (srv *GenericService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns a channel which is closed when the initial state of the service is received.
func (srv *GenericService) Ready() <-chan struct{} {
	return srv.wait.Wait()
}

func (srv *GenericService) HasBool(id string) bool {
	if srv == nil {
		return false
	}
	_, found := srv.bools[id]
	return found
}

func (srv *GenericService) Bool(id string) bool {
	if srv == nil {
		return false
	}
	if v, found := srv.bools[id]; found {
		return v
	}
	return false
}

func (srv *GenericService) SetBool(ctx context.Context, id string, value bool) error {
	if srv == nil {
		return fmt.Errorf("service not initialised")
	}
	return srv.requester(
		ctx,
		&clientsapi.ActionRequest{
			ServiceId: srv.id,
			Values: []*clientsapi.Value{
				{
					Id: id,
					Bool: &clientsapi.BoolValue{
						Value: value,
					},
				},
			},
		},
		func(resp *clientsapi.ActionResponse) {
		},
	)
}

func (srv *GenericService) HasFloat(id string) bool {
	if srv == nil {
		return false
	}
	_, found := srv.floats[id]
	return found
}

func (srv *GenericService) Float(id string) float64 {
	if srv == nil {
		return 0.0
	}
	if v, found := srv.floats[id]; found {
		return v
	}
	return 0.0
}

func (srv *GenericService) SetFloat(ctx context.Context, id string, value float64) error {
	if srv == nil {
		return fmt.Errorf("service not initialised")
	}
	return srv.requester(
		ctx,
		&clientsapi.ActionRequest{
			ServiceId: srv.id,
			Values: []*clientsapi.Value{
				{
					Id: id,
					Float: &clientsapi.FloatValue{
						Value: value,
					},
				},
			},
		},
		func(resp *clientsapi.ActionResponse) {
		},
	)
}
