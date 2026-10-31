package attributes

import (
	"fmt"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/schema"
	"github.com/jimjibone/woodhouse-api/values"
)

// FromSchema creates the attribute a schema field describes, for a service
// with the given features enabled. The concrete type follows the field type
// (*Bool for bool, *List for list, ...). A required field is Required,
// anything else Optional (only sent once set).
func FromSchema(f *schema.Field, features []string) Attribute {
	perms := values.Perms(f.Perms)
	optional := Optional
	if f.Required {
		optional = Required
	}
	spec := values.SpecForFeatures(f, features)
	switch f.Type {
	case schema.TypeBool:
		return NewBool(f.Name, perms, optional)
	case schema.TypeInt:
		return NewInt(f.Name, perms, optional, spec.GetInt())
	case schema.TypeFloat:
		return NewFloat(f.Name, perms, optional, spec.GetFloat())
	case schema.TypeText:
		return NewText(f.Name, perms, optional)
	case schema.TypeDuration:
		return NewDuration(f.Name, perms, optional, spec.GetDuration())
	case schema.TypeTime:
		return NewTime(f.Name, perms, optional)
	case schema.TypeColor:
		c := spec.GetColor()
		return NewColor(f.Name, perms, optional, c.GetHueSat(), c.GetXy())
	case schema.TypeEnum:
		return NewEnum(f.Name, perms, optional, spec.GetEnum().GetOptions()...)
	case schema.TypeImage:
		return NewImage(f.Name, optional)
	case schema.TypeList:
		return NewList(f.Name, perms, optional, spec.GetList())
	case schema.TypeStruct:
		return NewStruct(f.Name, perms, optional, spec.GetStruct())
	}
	panic(fmt.Sprintf("attributes: unknown schema type %q for %q", f.Type, f.Name))
}

// Static checks that every schema type has a matching spec kind.
var _ = map[schema.Type]any{
	schema.TypeBool:     (*clientsapi.Spec_Bool)(nil),
	schema.TypeInt:      (*clientsapi.Spec_Int)(nil),
	schema.TypeFloat:    (*clientsapi.Spec_Float)(nil),
	schema.TypeText:     (*clientsapi.Spec_Text)(nil),
	schema.TypeDuration: (*clientsapi.Spec_Duration)(nil),
	schema.TypeTime:     (*clientsapi.Spec_Time)(nil),
	schema.TypeColor:    (*clientsapi.Spec_Color)(nil),
	schema.TypeEnum:     (*clientsapi.Spec_Enum)(nil),
	schema.TypeImage:    (*clientsapi.Spec_Image)(nil),
	schema.TypeList:     (*clientsapi.Spec_List)(nil),
	schema.TypeStruct:   (*clientsapi.Spec_Struct)(nil),
}
