package ast

import (
	"fmt"
	"sort"
	"strings"
	"yangly/casing"
)

type TSType string

const (
	TS_STRING  TSType = "string"
	TS_NUMBER         = "number"
	TS_BIGINT         = "bigint"
	TS_BOOL           = "boolean"
	TS_ANY            = "any"
	TS_UNKNOWN        = "unknown"
	TS_NEVER          = "never"
)

type Stringable interface {
	String() string
}

type Type interface {
	String() string
	// TODO: It would be preferable to implement WriteTo(w io.Writer)
	// TODO:
	// SetComment(v string)
}

type Primitive struct {
	t TSType
}

func (p Primitive) String() string {
	// Must format value to cast to string from TSType(string)
	return fmt.Sprintf("%s", p.t)
}

type Array struct {
	t Type
}

func (a Array) String() string {
	return fmt.Sprintf("Array<%s>", a.t.String())
}

type Union struct {
	opts []string
}

func (u Union) String() string {
	return strings.Join(sort.StringSlice(u.opts), " | ")
}

type Record struct {
	key    Type
	fields []Field
}

func (r *Record) SetKeyType(t Type) {
	r.key = t
}

func (r *Record) AddField(field Field) {
	r.fields = append(r.fields, field)
}

func (r Record) String() string {
	if len(r.fields) == 0 {
		return fmt.Sprintf("Record<%s, %s>", r.key.String(), TS_ANY)
	}

	var members []string
	for _, f := range r.fields {
		members = append(members, f.String())
	}

	key := r.key.String()
	if key == "string" {
		// Format as object literal
		return fmt.Sprintf("{\n%s\n}", strings.Join(members, " "))
	}
	return fmt.Sprintf("Record<%s, {\n%s\n}>", key, strings.Join(members, "\n"))
}

type Field struct {
	name     string
	value    Type
	nullable bool
}

func (f Field) String() string {
	var nullableFlag string
	if f.nullable {
		nullableFlag = "?"
	}

	if strings.Contains(f.name, "-") {
		// quote fields containing dashes
		return fmt.Sprintf("\"%s\"%s: %s;\n", f.name, nullableFlag, f.value.String())
	} else {
		return fmt.Sprintf("%s%s: %s;\n", f.name, nullableFlag, f.value.String())
	}
}

type Interface struct {
	name   string
	fields []Field
}

func (i *Interface) AddField(field Field) {
	i.fields = append(i.fields, field)
}

func (i *Interface) FieldCount() int {
	return len(i.fields)
}

func (i Interface) String() string {
	var members []string
	for _, f := range i.fields {
		members = append(members, f.String())
	}

	return fmt.Sprintf("export interface %s {\n%s\n}", casing.ToPascalCase(i.name), strings.Join(members, "\n"))
}
