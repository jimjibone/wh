package services

import (
	"fmt"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

func PrettyServices(pad string, srvs []*clientsapi.Service) string {
	str := ""
	for i, srv := range srvs {
		if i > 0 {
			str += "\n"
		}
		str += PrettyService(pad, srv)
	}
	return str
}

func PrettyService(pad string, srv *clientsapi.Service) string {
	str := fmt.Sprintf("%s- srv id:%q, typ:%q, alias:%q, attrs:%d", pad, srv.GetId(), srv.GetTyp(), srv.GetAlias(), len(srv.Attrs))
	for _, attr := range srv.Attrs {
		str += fmt.Sprintf("\n%s  - attr %s", pad, attr)
	}
	return str
}
