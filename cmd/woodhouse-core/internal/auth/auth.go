package auth

import (
	"slices"
	"strings"
)

func IsUserMethod(method string) bool {
	return strings.HasPrefix(method, "/woodhouse.api.v1.clients.UserService/")
}

// roleMap is a map of method names (keys) and the list of roles which are allowed to access them.
var roleMap = map[string][]Role{
	"/woodhouse.api.v1.clients.UserService/GetClients":            {AdminRole, UserRole},
	"/woodhouse.api.v1.clients.UserService/ClientsStream":         {AdminRole, UserRole},
	"/woodhouse.api.v1.clients.UserService/PairingRequestsStream": {AdminRole},
	"/woodhouse.api.v1.clients.UserService/ApprovePairing":        {AdminRole},
	"/woodhouse.api.v1.clients.UserService/DenyPairing":           {AdminRole},
	"/woodhouse.api.v1.clients.UserService/UnpairClient":          {AdminRole},
	"/woodhouse.api.v1.clients.UserService/BlockClient":           {AdminRole},
	"/woodhouse.api.v1.clients.UserService/UnblockClient":         {AdminRole},
	"/woodhouse.api.v1.clients.UserService/ForgetClient":          {AdminRole},
	"/woodhouse.api.v1.clients.UserService/GetDevices":            {AdminRole, UserRole},
	"/woodhouse.api.v1.clients.UserService/DevicesStream":         {AdminRole, UserRole},
	"/woodhouse.api.v1.clients.UserService/FavoritesStream":       {AdminRole, UserRole},
	"/woodhouse.api.v1.clients.UserService/AddFavorite":           {AdminRole, UserRole},
	"/woodhouse.api.v1.clients.UserService/RemoveFavorite":        {AdminRole, UserRole},
	"/woodhouse.api.v1.clients.UserService/GroupsStream":          {AdminRole},
	"/woodhouse.api.v1.clients.UserService/AddGroup":              {AdminRole, UserRole},
	"/woodhouse.api.v1.clients.UserService/UpdateGroup":           {AdminRole, UserRole},
	"/woodhouse.api.v1.clients.UserService/RemoveGroup":           {AdminRole, UserRole},
	"/woodhouse.api.v1.clients.UserService/SendAction":            {AdminRole, UserRole},
	"/woodhouse.api.v1.clients.UserService/SendImageRequest":      {AdminRole, UserRole},
	"/woodhouse.api.v1.clients.UserService/UsersStream":           {AdminRole, UserRole},
	"/woodhouse.api.v1.clients.UserService/AddUser":               {AdminRole},
	"/woodhouse.api.v1.clients.UserService/UpdateUser":            {AdminRole, UserRole},
	"/woodhouse.api.v1.clients.UserService/RemoveUser":            {AdminRole},
}

func IsUserAuthorised(method string, role Role) bool {
	if roles, found := roleMap[method]; found {
		return slices.Contains(roles, role)
	}
	return false
}
