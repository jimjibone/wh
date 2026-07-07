package services

import (
	"strings"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"
)

// DefaultServiceID returns the default service ID for a given service type.
// Returns an empty string if there is no default ID for the service type.
func DefaultServiceID(serviceType clientsapi.Service_ServiceType) string {
	if id, ok := clientsapi.Service_ServiceType_name[int32(serviceType)]; ok {
		return strings.ToLower(id)
	}
	return ""
}
