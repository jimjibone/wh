package apitools

import clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"

func RequiresAuth(method string) bool {
	switch method {
	case
		clientsapi.AuthService_Pair_FullMethodName,
		clientsapi.AuthService_Refresh_FullMethodName,
		clientsapi.AuthService_Ping_FullMethodName:
		return false
	default:
		return true
	}
}
