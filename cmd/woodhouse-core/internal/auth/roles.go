package auth

import "fmt"

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
