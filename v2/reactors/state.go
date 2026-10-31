package reactors

import (
	"slices"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"google.golang.org/protobuf/proto"
)

// serviceKey addresses a service of a device.
type serviceKey struct {
	device, endpoint, service string
}

// touched is a service an update changed or removed.
type touched struct {
	key     serviceKey
	changed bool // an attribute value, the spec, features or the presence changed
}

// deviceState mirrors a device's state from the update stream: full states
// replace it, partial updates merge into it. Events are not kept.
type deviceState struct {
	pb *clientsapi.Device
}

func (st *deviceState) service(endpointID, serviceID string) *clientsapi.Service {
	if st == nil || st.pb == nil {
		return nil
	}
	for _, ep := range st.pb.GetEndpoints() {
		if ep.GetId() != endpointID {
			continue
		}
		for _, srv := range ep.GetServices() {
			if srv.GetId() == serviceID {
				return srv
			}
		}
	}
	return nil
}

func (st *deviceState) endpoint(id string) *clientsapi.Endpoint {
	if st == nil || st.pb == nil {
		return nil
	}
	for _, ep := range st.pb.GetEndpoints() {
		if ep.GetId() == id {
			return ep
		}
	}
	return nil
}

// apply merges an update and reports the services it touched.
func (st *deviceState) apply(update *clientsapi.Device) []touched {
	update = proto.Clone(update).(*clientsapi.Device)
	update.Events = nil
	key := func(ep, srv string) serviceKey { return serviceKey{update.GetId(), ep, srv} }

	if update.GetFullState() || st.pb == nil {
		old := st.pb
		st.pb = update
		var out []touched
		for _, ep := range update.GetEndpoints() {
			for _, srv := range ep.GetServices() {
				prev := (&deviceState{pb: old}).service(ep.GetId(), srv.GetId())
				out = append(out, touched{key(ep.GetId(), srv.GetId()), !proto.Equal(prev, srv)})
			}
		}
		// Services the full state no longer has.
		for _, ep := range old.GetEndpoints() {
			for _, srv := range ep.GetServices() {
				if st.service(ep.GetId(), srv.GetId()) == nil {
					out = append(out, touched{key(ep.GetId(), srv.GetId()), true})
				}
			}
		}
		return out
	}

	if update.GetTyp() != clientsapi.Device_UNDEFINED {
		st.pb.Typ = update.GetTyp()
	}
	var out []touched
	for _, upEp := range update.GetEndpoints() {
		ep := st.endpoint(upEp.GetId())
		if ep == nil {
			ep = &clientsapi.Endpoint{Id: upEp.GetId()}
			st.pb.Endpoints = append(st.pb.Endpoints, ep)
		}
		if upEp.GetLabel() != "" {
			ep.Label = upEp.GetLabel()
		}
		for _, upSrv := range upEp.GetServices() {
			srv := st.service(ep.GetId(), upSrv.GetId())
			if srv == nil {
				ep.Services = append(ep.Services, upSrv)
				out = append(out, touched{key(ep.GetId(), upSrv.GetId()), true})
				continue
			}
			out = append(out, touched{key(ep.GetId(), upSrv.GetId()), mergeService(srv, upSrv)})
		}
	}
	return out
}

// mergeService merges a partial service update and reports whether anything
// visible changed. An attribute update without a spec keeps the known spec
// and perms and takes the value and time.
func mergeService(srv, up *clientsapi.Service) bool {
	changed := false
	if up.GetAlias() != srv.GetAlias() {
		srv.Alias = up.GetAlias()
		changed = true
	}
	if len(up.GetFeatures()) > 0 && !slices.Equal(up.GetFeatures(), srv.GetFeatures()) {
		srv.Features = up.GetFeatures()
		changed = true
	}
	if len(up.GetCommands()) > 0 {
		srv.Commands = up.GetCommands()
	}
	if len(up.GetEvents()) > 0 {
		srv.Events = up.GetEvents()
	}
	for _, a := range up.GetAttrs() {
		i := slices.IndexFunc(srv.Attrs, func(x *clientsapi.Attribute) bool { return x.GetId() == a.GetId() })
		if i < 0 {
			srv.Attrs = append(srv.Attrs, a)
			changed = true
			continue
		}
		cur := srv.Attrs[i]
		if !proto.Equal(cur.GetValue(), a.GetValue()) {
			changed = true
		}
		if a.GetSpec() != nil {
			if !proto.Equal(cur.GetSpec(), a.GetSpec()) || cur.GetPerms() != a.GetPerms() {
				changed = true
			}
			srv.Attrs[i] = a
			continue
		}
		cur.Value = a.GetValue()
		cur.Time = a.GetTime()
	}
	return changed
}

func attribute(srv *clientsapi.Service, id string) *clientsapi.Attribute {
	for _, a := range srv.GetAttrs() {
		if a.GetId() == id {
			return a
		}
	}
	return nil
}
