package main

import (
	"fmt"
	"strconv"
	"strings"

	clientsapi "github.com/jimjibone/woodhouse-api/go/v2/clients"
	"github.com/jimjibone/woodhouse-api/schema"
	"github.com/jimjibone/woodhouse-api/schema/schemaspec"
)

// The generated code builds specs as Go literals computed from the schema
// here, at generate time, so nothing parses the schema at run time.

// featureCond returns a Go condition for "any of the features is enabled",
// or "" when the field names none.
func featureCond(svc string, features []string) string {
	var parts []string
	for _, f := range features {
		parts = append(parts, fmt.Sprintf("srv.HasFeature(%sFeature%s)", svc, camel(f)))
	}
	return strings.Join(parts, " || ")
}

// colorArgs returns the hue/sat and xy arguments of a color field: tied to
// features, they follow the "hs" and "xy" features at run time.
func colorArgs(svc string, f *schema.Field) (hueSat, xy string) {
	if len(f.Features) == 0 {
		return "true", "true"
	}
	return fmt.Sprintf("srv.HasFeature(%sFeatureHS)", svc), fmt.Sprintf("srv.HasFeature(%sFeatureXY)", svc)
}

// specExpr returns a Go expression of the field's *clientsapi.Spec.
func specExpr(svc string, f *schema.Field) string {
	if f.Type == schema.TypeColor {
		hs, xy := colorArgs(svc, f)
		return fmt.Sprintf("&clientsapi.Spec{Kind: &clientsapi.Spec_Color{Color: &clientsapi.ColorSpec{HueSat: %s, Xy: %s}}}", hs, xy)
	}
	return specLit(schemaspec.SpecFromSchema(f))
}

// fieldSpecExpr returns a Go expression of the field's *clientsapi.FieldSpec.
func fieldSpecExpr(svc string, f *schema.Field) string {
	return fmt.Sprintf("&clientsapi.FieldSpec{Id: %q, Spec: %s, Required: %t}", f.Name, specExpr(svc, f), f.Required)
}

// specLit returns a Go literal of a spec built without feature context.
func specLit(spec *clientsapi.Spec) string {
	switch k := spec.Kind.(type) {
	case *clientsapi.Spec_Bool:
		return "&clientsapi.Spec{Kind: &clientsapi.Spec_Bool{Bool: &clientsapi.BoolSpec{}}}"
	case *clientsapi.Spec_Int:
		return "&clientsapi.Spec{Kind: &clientsapi.Spec_Int{Int: " + intSpecLit(k.Int) + "}}"
	case *clientsapi.Spec_Float:
		return "&clientsapi.Spec{Kind: &clientsapi.Spec_Float{Float: " + floatSpecLit(k.Float) + "}}"
	case *clientsapi.Spec_Text:
		return "&clientsapi.Spec{Kind: &clientsapi.Spec_Text{Text: &clientsapi.TextSpec{}}}"
	case *clientsapi.Spec_Duration:
		return "&clientsapi.Spec{Kind: &clientsapi.Spec_Duration{Duration: " + durationSpecLit(k.Duration) + "}}"
	case *clientsapi.Spec_Time:
		return "&clientsapi.Spec{Kind: &clientsapi.Spec_Time{Time: &clientsapi.TimeSpec{}}}"
	case *clientsapi.Spec_Color:
		return fmt.Sprintf("&clientsapi.Spec{Kind: &clientsapi.Spec_Color{Color: &clientsapi.ColorSpec{HueSat: %t, Xy: %t}}}", k.Color.HueSat, k.Color.Xy)
	case *clientsapi.Spec_Enum:
		return "&clientsapi.Spec{Kind: &clientsapi.Spec_Enum{Enum: &clientsapi.EnumSpec{Options: " + stringsLit(k.Enum.Options) + "}}}"
	case *clientsapi.Spec_Image:
		return "&clientsapi.Spec{Kind: &clientsapi.Spec_Image{Image: &clientsapi.ImageSpec{}}}"
	case *clientsapi.Spec_List:
		return "&clientsapi.Spec{Kind: &clientsapi.Spec_List{List: " + listSpecLit(k.List) + "}}"
	case *clientsapi.Spec_Struct:
		return "&clientsapi.Spec{Kind: &clientsapi.Spec_Struct{Struct: " + structSpecLit(k.Struct) + "}}"
	}
	panic(fmt.Sprintf("whgen: unknown spec kind %T", spec.Kind))
}

func intSpecLit(s *clientsapi.IntSpec) string {
	var parts []string
	if s.Min != nil {
		parts = append(parts, "Min: ptr[int64]("+strconv.FormatInt(*s.Min, 10)+")")
	}
	if s.Max != nil {
		parts = append(parts, "Max: ptr[int64]("+strconv.FormatInt(*s.Max, 10)+")")
	}
	if s.Step != nil {
		parts = append(parts, "Step: ptr[uint64]("+strconv.FormatUint(*s.Step, 10)+")")
	}
	return "&clientsapi.IntSpec{" + strings.Join(append(parts, unitLit(s.Unit)...), ", ") + "}"
}

func floatSpecLit(s *clientsapi.FloatSpec) string {
	var parts []string
	if s.Min != nil {
		parts = append(parts, "Min: ptr[float64]("+strconv.FormatFloat(*s.Min, 'g', -1, 64)+")")
	}
	if s.Max != nil {
		parts = append(parts, "Max: ptr[float64]("+strconv.FormatFloat(*s.Max, 'g', -1, 64)+")")
	}
	if s.Step != nil {
		parts = append(parts, "Step: ptr[float64]("+strconv.FormatFloat(*s.Step, 'g', -1, 64)+")")
	}
	return "&clientsapi.FloatSpec{" + strings.Join(append(parts, unitLit(s.Unit)...), ", ") + "}"
}

// durationSpecLit writes durations in milliseconds, as on the wire.
func durationSpecLit(s *clientsapi.DurationSpec) string {
	var parts []string
	if s.Min != nil {
		parts = append(parts, "Min: ptr[int64]("+strconv.FormatInt(*s.Min, 10)+")")
	}
	if s.Max != nil {
		parts = append(parts, "Max: ptr[int64]("+strconv.FormatInt(*s.Max, 10)+")")
	}
	if s.Step != nil {
		parts = append(parts, "Step: ptr[uint64]("+strconv.FormatUint(*s.Step, 10)+")")
	}
	return "&clientsapi.DurationSpec{" + strings.Join(parts, ", ") + "}"
}

func listSpecLit(s *clientsapi.ListSpec) string {
	parts := []string{"Item: " + specLit(s.Item)}
	if s.MinItems != nil {
		parts = append(parts, "MinItems: ptr[uint32]("+strconv.FormatUint(uint64(*s.MinItems), 10)+")")
	}
	if s.MaxItems != nil {
		parts = append(parts, "MaxItems: ptr[uint32]("+strconv.FormatUint(uint64(*s.MaxItems), 10)+")")
	}
	return "&clientsapi.ListSpec{" + strings.Join(parts, ", ") + "}"
}

func structSpecLit(s *clientsapi.StructSpec) string {
	var fields []string
	for _, f := range s.Fields {
		fields = append(fields, fmt.Sprintf("{Id: %q, Spec: %s, Required: %t}", f.Id, specLit(f.Spec), f.Required))
	}
	return "&clientsapi.StructSpec{Fields: []*clientsapi.FieldSpec{" + strings.Join(fields, ", ") + "}}"
}

func unitLit(u clientsapi.Unit) []string {
	if u == clientsapi.Unit_UNIT_UNDEFINED {
		return nil
	}
	return []string{"Unit: clientsapi.Unit_" + u.String()}
}

func stringsLit(ss []string) string {
	if len(ss) == 0 {
		return "nil"
	}
	var q []string
	for _, s := range ss {
		q = append(q, strconv.Quote(s))
	}
	return "[]string{" + strings.Join(q, ", ") + "}"
}

// permsLit returns the Permissions constant for schema perms.
func permsLit(p schema.Perms) string {
	if p == schema.PermsRW {
		return "clientsapi.Permissions_PERM_READWRITE"
	}
	return "clientsapi.Permissions_PERM_READONLY"
}

// attrCtor returns the attributes.New... call that builds an attribute.
func attrCtor(svc string, a *schema.Field) string {
	opt := "attributes.Optional"
	if a.Required {
		opt = "attributes.Required"
	}
	perms := permsLit(a.Perms)
	head := fmt.Sprintf("%q, %s, %s", a.Name, perms, opt)
	spec := schemaspec.SpecFromSchema(a)
	switch a.Type {
	case schema.TypeBool:
		return "attributes.NewBool(" + head + ")"
	case schema.TypeText:
		return "attributes.NewText(" + head + ")"
	case schema.TypeTime:
		return "attributes.NewTime(" + head + ")"
	case schema.TypeInt:
		return "attributes.NewInt(" + head + ", " + intSpecLit(spec.GetInt()) + ")"
	case schema.TypeFloat:
		return "attributes.NewFloat(" + head + ", " + floatSpecLit(spec.GetFloat()) + ")"
	case schema.TypeDuration:
		return "attributes.NewDuration(" + head + ", " + durationSpecLit(spec.GetDuration()) + ")"
	case schema.TypeColor:
		hs, xy := colorArgs(svc, a)
		return "attributes.NewColor(" + head + ", " + hs + ", " + xy + ")"
	case schema.TypeEnum:
		args := head
		for _, o := range a.Options {
			args += ", " + strconv.Quote(o)
		}
		return "attributes.NewEnum(" + args + ")"
	case schema.TypeImage:
		return fmt.Sprintf("attributes.NewImage(%q, %s)", a.Name, opt)
	case schema.TypeList:
		return "attributes.NewList(" + head + ", " + listSpecLit(spec.GetList()) + ")"
	case schema.TypeStruct:
		return "attributes.NewStruct(" + head + ", " + structSpecLit(spec.GetStruct()) + ")"
	}
	panic(fmt.Sprintf("whgen: unknown schema type %q", a.Type))
}
