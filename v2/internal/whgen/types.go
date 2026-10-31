package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jimjibone/woodhouse-api/schema"
)

// structType is a Go struct generated for a schema struct, a list item struct,
// or the params, result or fields of a command or event. They live in package
// types and are shared by the bridge services and the reactor views.
type structType struct {
	name   string
	doc    string
	prefix string // name prefix for nested struct types
	fields []*schema.Field
	sig    string
}

// types maps schema fields to Go types and conversion expressions, and
// collects the struct types they need. qual qualifies the struct types and
// their Parse funcs: "" inside package types, "types." elsewhere.
type types struct {
	qual    string
	structs map[string]*structType
	order   []string
	err     error
}

func newTypes(qual string) *types {
	return &types{qual: qual, structs: map[string]*structType{}}
}

// collectStructs registers every struct type the schema needs, in a stable
// order, so each generator agrees on the names.
func collectStructs(t *types, s *schema.Schema) {
	for _, srv := range s.Services {
		if srv.Open {
			continue
		}
		name := camel(srv.Name)
		for _, a := range srv.Attributes {
			if a.Type == schema.TypeList || a.Type == schema.TypeStruct {
				t.goType(a, name+camel(a.Name))
			}
		}
		for _, c := range srv.Commands {
			if len(c.Params) > 0 {
				t.addStruct(paramsName(srv, c), fmt.Sprintf("are the params of %s.%s.", srv.Name, c.Name), name, c.Params)
			}
			if len(c.Result) > 0 {
				t.addStruct(resultName(srv, c), fmt.Sprintf("is the result of %s.%s.", srv.Name, c.Name), name, c.Result)
			}
		}
		for _, e := range srv.Events {
			if len(e.Fields) > 0 {
				t.addStruct(eventName(srv, e), fmt.Sprintf("are the fields of the %s.%s event.", srv.Name, e.Name), name, e.Fields)
			}
		}
	}
}

func paramsName(srv *schema.Service, c *schema.Command) string {
	return camel(srv.Name) + camel(c.Name) + "Params"
}

func resultName(srv *schema.Service, c *schema.Command) string {
	return camel(srv.Name) + camel(c.Name) + "Result"
}

func eventName(srv *schema.Service, e *schema.Event) string {
	return camel(srv.Name) + camel(e.Name) + "Event"
}

// optionalPtr reports whether an optional field of this type is a pointer.
// Lists and colors are already nil-able.
func optionalPtr(f *schema.Field) bool {
	return !f.Required && f.Type != schema.TypeList && f.Type != schema.TypeColor
}

// goType returns the Go type of a field's value. typeName names the struct
// type if the field (or its list item) is a struct.
func (t *types) goType(f *schema.Field, typeName string) string {
	switch f.Type {
	case schema.TypeBool:
		return "bool"
	case schema.TypeInt:
		return "int64"
	case schema.TypeFloat:
		return "float64"
	case schema.TypeText, schema.TypeEnum:
		return "string"
	case schema.TypeDuration:
		return "time.Duration"
	case schema.TypeTime:
		return "time.Time"
	case schema.TypeColor:
		return "*clientsapi.ColorValue"
	case schema.TypeList:
		return "[]" + t.goType(f.Item, typeName+"Item")
	case schema.TypeStruct:
		t.addStruct(typeName, "", typeName, f.Fields)
		return t.qual + typeName
	}
	t.fail("field %q: type %q has no Go type here", f.Name, f.Type)
	return "any"
}

// fieldType is goType with optional fields as pointers where needed.
func (t *types) fieldType(f *schema.Field, typeName string) string {
	gt := t.goType(f, typeName)
	if optionalPtr(f) {
		return "*" + gt
	}
	return gt
}

// toValue returns an expression converting the Go value expr to a
// *clientsapi.Value.
func (t *types) toValue(expr string, f *schema.Field, typeName string) string {
	switch f.Type {
	case schema.TypeBool:
		return "values.Bool(" + expr + ")"
	case schema.TypeInt:
		return "values.Int(" + expr + ")"
	case schema.TypeFloat:
		return "values.Float(" + expr + ")"
	case schema.TypeText:
		return "values.Text(" + expr + ")"
	case schema.TypeEnum:
		return "values.Enum(" + expr + ")"
	case schema.TypeDuration:
		return "values.Duration(" + expr + ")"
	case schema.TypeTime:
		return "values.Time(" + expr + ")"
	case schema.TypeColor:
		return "values.Color(" + expr + ".GetHueSat(), " + expr + ".GetXy())"
	case schema.TypeStruct:
		return expr + ".Value()"
	case schema.TypeList:
		item := t.toValue("it", f.Item, typeName+"Item")
		return "func() *clientsapi.Value { items := make([]*clientsapi.Value, 0, len(" + expr + ")); for _, it := range " + expr +
			" { items = append(items, " + item + ") }; return values.List(items...) }()"
	}
	t.fail("field %q: type %q cannot be converted", f.Name, f.Type)
	return "nil"
}

// fromValue returns an expression converting the *clientsapi.Value expr to
// the field's Go value type.
func (t *types) fromValue(expr string, f *schema.Field, typeName string) string {
	switch f.Type {
	case schema.TypeBool:
		return expr + ".GetBool()"
	case schema.TypeInt:
		return expr + ".GetInt()"
	case schema.TypeFloat:
		return expr + ".GetFloat()"
	case schema.TypeText:
		return expr + ".GetText()"
	case schema.TypeEnum:
		return expr + ".GetEnum()"
	case schema.TypeDuration:
		return t.qual + "DurationOf(" + expr + ")"
	case schema.TypeTime:
		return t.qual + "TimeOf(" + expr + ")"
	case schema.TypeColor:
		return expr + ".GetColor()"
	case schema.TypeStruct:
		return t.qual + "Parse" + typeName + "(" + expr + ".GetStruct().GetFields())"
	case schema.TypeList:
		itemType := t.goType(f.Item, typeName+"Item")
		item := t.fromValue("it", f.Item, typeName+"Item")
		return "func() []" + itemType + " { var out []" + itemType + "; for _, it := range " + expr +
			".GetList().GetItems() { out = append(out, " + item + ") }; return out }()"
	}
	t.fail("field %q: type %q cannot be converted", f.Name, f.Type)
	return "nil"
}

// addStruct registers a struct type. prefix names its nested struct types.
// A name registered twice must describe the same fields.
func (t *types) addStruct(name, doc, prefix string, fields []*schema.Field) {
	sig := signature(fields)
	if prev, ok := t.structs[name]; ok {
		if prev.sig != sig {
			t.fail("struct %s is generated twice with different fields", name)
		}
		if prev.doc == "" {
			prev.doc = doc
		}
		return
	}
	st := &structType{name: name, doc: doc, prefix: prefix, fields: fields, sig: sig}
	t.structs[name] = st
	t.order = append(t.order, name)
	for _, f := range fields {
		t.goType(f, prefix+camel(f.Name)) // register nested structs
	}
}

func signature(fields []*schema.Field) string {
	var parts []string
	for _, f := range fields {
		parts = append(parts, fmt.Sprintf("%s:%s:%t:%s", f.Name, f.Type, f.Required, fieldSig(f)))
	}
	return strings.Join(parts, ";")
}

func fieldSig(f *schema.Field) string {
	switch f.Type {
	case schema.TypeList:
		return "[" + string(f.Item.Type) + fieldSig(f.Item) + "]"
	case schema.TypeStruct:
		return "{" + signature(f.Fields) + "}"
	case schema.TypeEnum:
		opts := append([]string(nil), f.Options...)
		sort.Strings(opts)
		return strings.Join(opts, ",")
	}
	return ""
}

func (t *types) fail(format string, args ...any) {
	if t.err == nil {
		t.err = fmt.Errorf(format, args...)
	}
}
