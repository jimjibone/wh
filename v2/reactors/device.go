package reactors

import (
	"time"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/values"
	"google.golang.org/protobuf/proto"
)

// DeviceView is a reactor's view of a whole device: its name, online state
// and the endpoints and services it reports.
type DeviceView struct {
	rc *Reactor
	id string
}

// EndpointInfo describes one endpoint of a device.
type EndpointInfo struct {
	ID       string
	Label    string
	Services []ServiceInfo
}

// ServiceInfo identifies one service of an endpoint.
type ServiceInfo struct {
	ID       string
	Typ      clientsapi.Service_ServiceType
	Alias    string
	Features []string
}

// Device returns a view of the device. It is valid before the device is known.
func (rc *Reactor) Device(id string) *DeviceView {
	if id == "" {
		panic("device id must be defined")
	}
	return &DeviceView{rc: rc, id: id}
}

func (dv *DeviceView) ID() string { return dv.id }

// Exists reports whether the server has sent the device.
func (dv *DeviceView) Exists() bool {
	dv.rc.mu.RLock()
	defer dv.rc.mu.RUnlock()
	return dv.rc.devices[dv.id] != nil
}

func (dv *DeviceView) rootValue(service, attr string) *clientsapi.Value {
	dv.rc.mu.RLock()
	defer dv.rc.mu.RUnlock()
	return attribute(dv.rc.devices[dv.id].service("0", service), attr).GetValue()
}

// Name returns the name from the info service.
func (dv *DeviceView) Name() string { return dv.rootValue("info", "name").GetText() }

// Online returns the online state from the online service.
func (dv *DeviceView) Online() bool { return dv.rootValue("online", "online").GetBool() }

// LastSeen returns the last seen time from the online service.
func (dv *DeviceView) LastSeen() time.Time {
	t, _ := values.AsTime(dv.rootValue("online", "last_seen"))
	return t
}

// Typ returns the device type.
func (dv *DeviceView) Typ() clientsapi.Device_DeviceType {
	dv.rc.mu.RLock()
	defer dv.rc.mu.RUnlock()
	return dv.rc.devices[dv.id].pbOrNil().GetTyp()
}

// Endpoints lists the device's endpoints and their services.
func (dv *DeviceView) Endpoints() []EndpointInfo {
	dv.rc.mu.RLock()
	defer dv.rc.mu.RUnlock()
	var out []EndpointInfo
	for _, ep := range dv.rc.devices[dv.id].pbOrNil().GetEndpoints() {
		info := EndpointInfo{ID: ep.GetId(), Label: ep.GetLabel()}
		for _, srv := range ep.GetServices() {
			info.Services = append(info.Services, ServiceInfo{ID: srv.GetId(), Typ: srv.GetTyp(), Alias: srv.GetAlias(), Features: srv.GetFeatures()})
		}
		out = append(out, info)
	}
	return out
}

// Pb returns a copy of the mirrored device state, or nil.
func (dv *DeviceView) Pb() *clientsapi.Device {
	dv.rc.mu.RLock()
	defer dv.rc.mu.RUnlock()
	pb := dv.rc.devices[dv.id].pbOrNil()
	if pb == nil {
		return nil
	}
	return proto.Clone(pb).(*clientsapi.Device)
}

func (st *deviceState) pbOrNil() *clientsapi.Device {
	if st == nil {
		return nil
	}
	return st.pb
}
