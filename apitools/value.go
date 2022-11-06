package apitools

import (
	"fmt"

	api "github.com/jimjibone/woodhouse-4/api/go"
)

func ValueAs(value *api.DeviceValue, out interface{}) bool {
	switch outVal := out.(type) {
	case *api.BoolValue:
		if value.Bool != nil {
			outVal.Value = value.Bool.Value
			return true
		}

	case *api.NumberValue:
		if value.Number != nil {
			outVal.Value = value.Number.Value
			return true
		}

	case *api.TextValue:
		if value.Text != nil {
			outVal.Value = value.Text.Value
			return true
		}

	default:
		panic(fmt.Sprintf("unsupported output type: %+v", out))
	}

	return false
}

func ValueFrom(name string, value interface{}) *api.DeviceValue {
	out := &api.DeviceValue{Name: name}

	switch val := value.(type) {
	case *api.BoolValue:
		out.Bool = val

	case *api.NumberValue:
		out.Number = val

	case *api.TextValue:
		out.Text = val

	default:
		panic(fmt.Sprintf("unsupported value type: %+v", value))
	}

	return nil
}
