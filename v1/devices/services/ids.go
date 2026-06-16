package services

import clientsapi "github.com/jimjibone/woodhouse-api/go/v1/clients"

var serviceTypeToID map[clientsapi.Service_ServiceType]string

func registerDefaultServiceID(serviceType clientsapi.Service_ServiceType, id string) {
	if serviceTypeToID == nil {
		serviceTypeToID = make(map[clientsapi.Service_ServiceType]string)
	}
	serviceTypeToID[serviceType] = id
}

// DefaultServiceID returns the default service ID for a given service type.
// Returns an empty string if there is no default ID for the service type.
func DefaultServiceID(serviceType clientsapi.Service_ServiceType) string {
	if id, ok := serviceTypeToID[serviceType]; ok {
		return id
	}
	return ""
}
