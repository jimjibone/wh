package auth

import (
	"fmt"

	clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"
)

type Role int

const (
	NoAuthRole Role = iota
	AdminRole
	UserRole
)

func (r Role) String() string {
	switch r {
	case NoAuthRole:
		return "noauth"
	case AdminRole:
		return "admin"
	case UserRole:
		return "user"
	}
	return "<UNIMPLEMENTED>"
}

func (r Role) MarshalJSON() ([]byte, error) {
	switch r {
	case NoAuthRole:
		return []byte(`"noauth"`), nil
	case AdminRole:
		return []byte(`"admin"`), nil
	case UserRole:
		return []byte(`"user"`), nil
	}
	return nil, fmt.Errorf("unimplemented")
}

func (r *Role) UnmarshalJSON(p []byte) error {
	switch string(p) {
	case `"noauth"`:
		*r = NoAuthRole
	case `"admin"`:
		*r = AdminRole
	case `"user"`:
		*r = UserRole
	default:
		return fmt.Errorf("unknown")
	}
	return nil
}

func (r Role) Pb() clientsapi.UserRole {
	switch r {
	case NoAuthRole:
		return clientsapi.UserRole_USER_ROLE_UNDEFINED
	case AdminRole:
		return clientsapi.UserRole_USER_ROLE_ADMIN
	case UserRole:
		return clientsapi.UserRole_USER_ROLE_USER
	}
	return clientsapi.UserRole_USER_ROLE_UNDEFINED
}

func RoleFromPb(pb clientsapi.UserRole) Role {
	switch pb {
	case clientsapi.UserRole_USER_ROLE_UNDEFINED:
		return NoAuthRole
	case clientsapi.UserRole_USER_ROLE_ADMIN:
		return AdminRole
	case clientsapi.UserRole_USER_ROLE_USER:
		return UserRole
	default:
	}
	return NoAuthRole
}
