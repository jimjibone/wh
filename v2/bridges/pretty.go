package bridges

import (
	"fmt"
	"strings"

	"github.com/jimjibone/wh/v2/bridges/services"
	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
)

// PrettyDevice formats a device update for debug logs.
func PrettyDevice(pad string, dev *clientsapi.Device) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s- device id:%q, typ:%s, full_state:%t, endpoints:%d, events:%d", pad, dev.GetId(), dev.GetTyp(), dev.GetFullState(), len(dev.GetEndpoints()), len(dev.GetEvents()))
	for _, ep := range dev.GetEndpoints() {
		fmt.Fprintf(&b, "\n%s  - endpoint id:%q, label:%q, services:%d", pad, ep.GetId(), ep.GetLabel(), len(ep.GetServices()))
		if len(ep.GetServices()) > 0 {
			b.WriteString("\n" + services.PrettyServices(pad+"    ", ep.GetServices()))
		}
	}
	for _, e := range dev.GetEvents() {
		fmt.Fprintf(&b, "\n%s  - event %s", pad, e)
	}
	return b.String()
}
