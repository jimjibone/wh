package reactors

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type LightbulbService struct {
	requester  Requester
	id         string
	on         bool
	brightness *int64
	saturation *int64
	hue        *float64
	colorTemp  *float64
	transition *time.Duration
}

func (srv *LightbulbService) handleUpdate(update *clientsapi.Service) bool {
	changed := false
	srv.id = update.GetId()
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
				srv.colorTemp = new(float64)
			}
			if *srv.colorTemp != attr.GetFloat().GetValue() {
				changed = true
				*srv.colorTemp = attr.GetFloat().GetValue()
			}
		}
	}
	return changed
}

func (srv *LightbulbService) On() bool {
	if srv == nil {
		return false
	}
	return srv.on
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

func (srv *LightbulbService) ColorTemp() float64 {
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
