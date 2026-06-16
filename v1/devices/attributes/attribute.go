package attributes

import clientsapi "github.com/jimjibone/woodhouse-4/api/go/v1/clients"

type OptionalType int

const (
	Optional OptionalType = iota
	Required
)

type Attribute interface {
	ID() string
	Perms() clientsapi.Permissions
	Optional() OptionalType
	IsSet() bool
	Push(push func(*clientsapi.Attribute))
	Pb() *clientsapi.Attribute
}
