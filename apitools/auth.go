package apitools

func RequiresAuth(method string) bool {
	switch method {
	case "/woodhouse.api.v1.clients.AuthService/Pair",
		"/woodhouse.api.v1.clients.AuthService/Refresh",
		"/woodhouse.api.v1.clients.AuthService/Ping":
		return false
	default:
		return true
	}
}
