package generator

import (
	"fmt"
	"sort"

	"github.com/untrustedmodders/go-plugify/manifest"
)

// TypeTables collects the prototypes and enums encountered while converting
// exported functions. Each is written to the manifest once and named from every
// use site, so a callback taken by twenty functions is described once.
//
// Two Go types in different packages can share a short name, and a name is what
// a reference resolves on, so a clash is recorded rather than silently letting
// one definition stand in for the other.
type TypeTables struct {
	prototypes map[string]*manifest.Method
	enums      map[string]*manifest.Enum
	clashes    []string
}

func newTypeTables() *TypeTables {
	return &TypeTables{
		prototypes: map[string]*manifest.Method{},
		enums:      map[string]*manifest.Enum{},
	}
}

// addPrototype files a callback signature and returns the name to refer to it by.
func (t *TypeTables) addPrototype(prototype *manifest.Method) string {
	existing, found := t.prototypes[prototype.Name]
	if !found {
		t.prototypes[prototype.Name] = prototype
		return prototype.Name
	}
	if !samePrototype(existing, prototype) {
		t.clash("prototype", prototype.Name)
	}
	return prototype.Name
}

// addEnum files an enumeration and returns the name to refer to it by.
func (t *TypeTables) addEnum(enum *manifest.Enum) string {
	existing, found := t.enums[enum.Name]
	if !found {
		t.enums[enum.Name] = enum
		return enum.Name
	}
	if !sameEnum(existing, enum) {
		t.clash("enum", enum.Name)
	}
	return enum.Name
}

func (t *TypeTables) clash(kind, name string) {
	clash := fmt.Sprintf("%s %s", kind, name)
	for _, seen := range t.clashes {
		if seen == clash {
			return
		}
	}
	t.clashes = append(t.clashes, clash)
}

// Err reports the name clashes found during conversion, if any.
func (t *TypeTables) Err() error {
	if len(t.clashes) == 0 {
		return nil
	}
	return fmt.Errorf(
		"conflicting definitions share a name, so a reference to one would be ambiguous: %v",
		t.clashes)
}

func (t *TypeTables) Prototypes() []*manifest.Method {
	if len(t.prototypes) == 0 {
		return nil
	}
	out := make([]*manifest.Method, 0, len(t.prototypes))
	for _, p := range t.prototypes {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (t *TypeTables) Enums() []*manifest.Enum {
	if len(t.enums) == 0 {
		return nil
	}
	out := make([]*manifest.Enum, 0, len(t.enums))
	for _, e := range t.enums {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ConvertToManifestMethods converts exported functions into manifest methods,
// filing every prototype and enum into the returned tables.
func ConvertToManifestMethods(funcs []ExportedFunction) ([]manifest.Method, *TypeTables) {
	tables := newTypeTables()

	methods := make([]manifest.Method, len(funcs))
	for i, f := range funcs {
		methods[i] = manifest.Method{
			Name:        f.ExportName,
			FuncName:    f.FuncName,
			Description: f.Description,
			ParamTypes:  tables.convertParams(f.Params),
			RetType:     tables.convertReturnType(ParamInfo{"", f.ReturnType, f.Description}),
		}
	}

	return methods, tables
}

func (t *TypeTables) convertParams(params []ParamInfo) []manifest.Property {
	result := make([]manifest.Property, len(params))
	for i, p := range params {
		result[i] = t.convertParamType(p, false)
	}
	return result
}

func (t *TypeTables) convertParamType(p ParamInfo, ignoreRef bool) manifest.Property {
	ty := p.Type

	if ty.IsFunc {
		// Function parameter. The signature is filed under its own name, and its
		// parameters are converted first so nested types are filed too.
		prototype := &manifest.Method{
			Name:        ty.FuncSig.Name,
			Description: ty.FuncSig.Description,
			FuncName:    "_",
			ParamTypes:  t.convertParams(ty.FuncSig.Params),
			RetType:     t.convertReturnType(ParamInfo{"", ty.FuncSig.Return, ty.FuncSig.Return.Description}),
		}

		return manifest.Property{
			Type:        "function",
			Name:        p.Name,
			Description: p.Description,
			Ref:         ty.IsRef && !ignoreRef,
			Prototype:   t.addPrototype(prototype),
		}
	}

	if ty.IsArray && ty.ElemType != nil && ty.ElemType.IsEnum {
		// Array of enum: the element carries the enum, the array carries the type.
		elem := t.convertParamType(ParamInfo{p.Name, *ty.ElemType, p.Description}, ignoreRef)

		return manifest.Property{
			Name:        p.Name,
			Description: p.Description,
			Type:        ty.TypeString,
			Ref:         ty.IsRef && !ignoreRef,
			Enum:        elem.Enum,
		}
	}

	if ty.IsEnum {
		values := make([]manifest.Value, len(ty.EnumValues))
		for i, v := range ty.EnumValues {
			values[i] = manifest.Value{
				Name:        v.Name,
				Description: v.Description,
				Value:       v.Value,
			}
		}

		return manifest.Property{
			Name:        p.Name,
			Description: p.Description,
			Type:        ty.TypeString,
			Ref:         ty.IsRef,
			Enum: t.addEnum(&manifest.Enum{
				Name:        ty.EnumTypeName,
				Description: ty.Description,
				Values:      values,
			}),
		}
	}

	if ty.IsAlias {
		return manifest.Property{
			Name:        p.Name,
			Description: p.Description,
			Type:        ty.TypeString,
			Ref:         ty.IsRef,
			Alias: &manifest.Alias{
				Name:        ty.EnumTypeName,
				Description: ty.Description,
			},
		}
	}

	return manifest.Property{
		Name:        p.Name,
		Description: p.Description,
		Type:        ty.TypeString,
		Ref:         ty.IsRef,
	}
}

func (t *TypeTables) convertReturnType(p ParamInfo) manifest.Property {
	return t.convertParamType(p, true)
}

// Definitions are compared on the parts that make up the type. Descriptions are
// left out: they mean nothing to a caller, and two spellings of the same type
// that differ only in wording are not a clash.

func sameProperty(a, b *manifest.Property) bool {
	return a.Type == b.Type && a.Ref == b.Ref &&
		a.Prototype == b.Prototype && a.Enum == b.Enum
}

func sameEnum(a, b *manifest.Enum) bool {
	if len(a.Values) != len(b.Values) {
		return false
	}
	for i := range a.Values {
		if a.Values[i].Name != b.Values[i].Name || a.Values[i].Value != b.Values[i].Value {
			return false
		}
	}
	return true
}

func samePrototype(a, b *manifest.Method) bool {
	if len(a.ParamTypes) != len(b.ParamTypes) {
		return false
	}
	if !sameProperty(&a.RetType, &b.RetType) {
		return false
	}
	for i := range a.ParamTypes {
		if !sameProperty(&a.ParamTypes[i], &b.ParamTypes[i]) {
			return false
		}
	}
	return true
}
