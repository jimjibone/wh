package apitools

import (
	api "github.com/jimjibone/woodhouse-4/api/go"
)

type BoolValue struct {
	Name  string
	Value bool
}

func (bv *BoolValue) GetValue() *api.DeviceValue {
	return &api.DeviceValue{
		Name: bv.Name,
		Bool: &api.BoolValue{
			Value: bv.Value,
		},
	}
}

func (bv *BoolValue) Parse(req *api.DeviceValue) (v bool, ok bool) {
	val := &api.BoolValue{}
	if ValueAs(req, val) {
		return val.Value, true
	}
	return val.Value, false
}

type NumberValue struct {
	Name  string
	Value float64
}

func (nv *NumberValue) GetValue() *api.DeviceValue {
	return &api.DeviceValue{
		Name: nv.Name,
		Number: &api.NumberValue{
			Value: nv.Value,
		},
	}
}

func (nv *NumberValue) Parse(req *api.DeviceValue) (v float64, ok bool) {
	val := &api.NumberValue{}
	if ValueAs(req, val) {
		return val.Value, true
	}
	return val.Value, false
}
