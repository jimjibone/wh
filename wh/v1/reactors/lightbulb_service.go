package reactors

import (
	"context"
	"fmt"
	"strings"
	"time"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type LightbulbService struct {
	id        string
	onUpdate  func(changed bool)
	requester requester
	exists    bool
	onliner

	on         bool
	brightness *int64
	saturation *int64
	hue        *float64
	colorTemp  *int64
	transition *time.Duration
}

type LightbulbRequest struct {
	On         *bool
	Brightness *int64
	Saturation *int64
	Hue        *float64
	ColorTemp  *int64
	Transition *time.Duration
}

func (l LightbulbRequest) String() string {
	var vals []string
	if l.On != nil {
		if *l.On {
			vals = append(vals, "on")
		} else {
			vals = append(vals, "off")
		}
	}
	if l.Brightness != nil {
		vals = append(vals, fmt.Sprintf("bri: %d%%", *l.Brightness))
	}
	if l.Hue != nil {
		vals = append(vals, fmt.Sprintf("hue: %.0f°", *l.Hue))
	}
	if l.Saturation != nil {
		vals = append(vals, fmt.Sprintf("sat: %d%%", *l.Saturation))
	}
	if l.ColorTemp != nil {
		vals = append(vals, fmt.Sprintf("ct: %d mireds", *l.ColorTemp))
	}
	if l.Transition != nil {
		vals = append(vals, fmt.Sprintf("dt: %s", *l.Transition))
	}
	return "{" + strings.Join(vals, ", ") + "}"
}

// Initialises the service.
func (srv *LightbulbService) init(serviceID string, requester requester) {
	srv.id = serviceID
	srv.requester = requester
}

// Handle the update. Returns true if the values changed.
func (srv *LightbulbService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	for _, attr := range update.Attrs {
		switch attr.GetId() {
		case "on":
			if srv.on != attr.GetBool().GetValue() {
				changed = true
				srv.on = attr.GetBool().GetValue()
			}

		case "brightness":
			if srv.brightness == nil {
				srv.brightness = new(int64)
			}
			if *srv.brightness != attr.GetInt().GetValue() {
				changed = true
				*srv.brightness = attr.GetInt().GetValue()
			}

		case "saturation":
			if srv.saturation == nil {
				srv.saturation = new(int64)
			}
			if *srv.saturation != attr.GetInt().GetValue() {
				changed = true
				*srv.saturation = attr.GetInt().GetValue()
			}

		case "hue":
			if srv.hue == nil {
				srv.hue = new(float64)
			}
			if *srv.hue != attr.GetFloat().GetValue() {
				changed = true
				*srv.hue = attr.GetFloat().GetValue()
			}

		case "color_temp":
			if srv.colorTemp == nil {
				srv.colorTemp = new(int64)
			}
			if *srv.colorTemp != attr.GetInt().GetValue() {
				changed = true
				*srv.colorTemp = attr.GetInt().GetValue()
			}
		}
	}
	if srv.onUpdate != nil {
		srv.onUpdate(changed)
	}
	srv.exists = true
	return changed
}

// Sets a handler to be called when the service is updated.
func (srv *LightbulbService) OnUpdate(handler func(changed bool)) {
	srv.onUpdate = handler
	srv.onliner.onUpdate = handler
}

// Returns whether the service exists or not. May be false until the client
// receives the initial state from the server.
func (srv *LightbulbService) Exists() bool {
	return srv.exists
}

func (srv *LightbulbService) Request(ctx context.Context, req LightbulbRequest, handler ...func(*clientsapi.ActionResponse)) error {
	if srv == nil {
		return fmt.Errorf("service not initialised")
	}

	var values []*clientsapi.Value
	if req.On != nil {
		values = append(values, &clientsapi.Value{
			Id: "on",
			Bool: &clientsapi.BoolValue{
				Value: *req.On,
			},
		})
	}
	if req.Brightness != nil {
		values = append(values, &clientsapi.Value{
			Id: "brightness",
			Int: &clientsapi.IntValue{
				Value: *req.Brightness,
			},
		})
	}
	if req.Saturation != nil {
		values = append(values, &clientsapi.Value{
			Id: "saturation",
			Int: &clientsapi.IntValue{
				Value: *req.Saturation,
			},
		})
	}
	if req.Hue != nil {
		values = append(values, &clientsapi.Value{
			Id: "hue",
			Float: &clientsapi.FloatValue{
				Value: *req.Hue,
			},
		})
	}
	if req.ColorTemp != nil {
		values = append(values, &clientsapi.Value{
			Id: "color_temp",
			Int: &clientsapi.IntValue{
				Value: *req.ColorTemp,
			},
		})
	}
	if req.Transition != nil {
		values = append(values, &clientsapi.Value{
			Id: "transition",
			Duration: &clientsapi.DurationValue{
				Value: req.Transition.Milliseconds(),
			},
		})
	}

	handlerFunc := func(*clientsapi.ActionResponse) {}
	if len(handler) > 0 {
		handlerFunc = handler[0]
	}
	return srv.requester(
		ctx,
		&clientsapi.ActionRequest{
			ServiceId: srv.id,
			Values:    values,
		},
		handlerFunc,
	)
}

func (srv *LightbulbService) On() bool {
	if srv == nil {
		return false
	}
	return srv.on
}

func (srv *LightbulbService) SetOn(ctx context.Context, on bool) error {
	return srv.Request(ctx, LightbulbRequest{On: &on}, func(ar *clientsapi.ActionResponse) {})
}

func (srv *LightbulbService) HasBrightness() bool {
	if srv == nil || srv.brightness == nil {
		return false
	}
	return true
}

func (srv *LightbulbService) Brightness() int64 {
	if srv == nil || srv.brightness == nil {
		return 0
	}
	return *srv.brightness
}

func (srv *LightbulbService) SetBrightness(ctx context.Context, brightness int64) error {
	return srv.Request(ctx, LightbulbRequest{Brightness: &brightness}, func(ar *clientsapi.ActionResponse) {})
}

func (srv *LightbulbService) HasSaturation() bool {
	if srv == nil || srv.saturation == nil {
		return false
	}
	return true
}

func (srv *LightbulbService) Saturation() int64 {
	if srv == nil || srv.saturation == nil {
		return 0
	}
	return *srv.saturation
}

func (srv *LightbulbService) HasHue() bool {
	if srv == nil || srv.hue == nil {
		return false
	}
	return true
}

func (srv *LightbulbService) Hue() float64 {
	if srv == nil || srv.hue == nil {
		return 0.0
	}
	return *srv.hue
}

func (srv *LightbulbService) HasColorTemp() bool {
	if srv == nil || srv.colorTemp == nil {
		return false
	}
	return true
}

func (srv *LightbulbService) ColorTemp() int64 {
	if srv == nil || srv.colorTemp == nil {
		return 0.0
	}
	return *srv.colorTemp
}

func (srv *LightbulbService) HasTransition() bool {
	if srv == nil || srv.transition == nil {
		return false
	}
	return true
}

func (srv *LightbulbService) Transition() time.Duration {
	if srv == nil || srv.transition == nil {
		return 0
	}
	return *srv.transition
}
