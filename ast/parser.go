package ast

import (
	"errors"
	"fmt"

	"github.com/sjwalker189/goyang/pkg/yang"
)

func NewParser(module *yang.Module) Parser {
	return Parser{
		yang:   module,
		strict: true,
	}
}

type Parser struct {
	yang   *yang.Module
	strict bool
}

func (p Parser) SetStrict(v bool) Parser {
	p.strict = true
	return p
}

// Parse a yang entry into a typescript type
func (sp *Parser) parseEntry(e *yang.Entry) Type {
	switch e.Kind {
	case yang.AnyXMLEntry:
		return Primitive{t: TS_STRING}
	case yang.AnyDataEntry:
		return sp.defaultType()
	case yang.LeafEntry:
		if e.IsLeafList() {
			// Leafs can be arrays of other types
			return Array{
				t: sp.parseYangType(e.Type),
			}
		}
		return sp.parseYangType(e.Type)
	case yang.DirectoryEntry:
		node := Record{
			key: Primitive{t: TS_STRING},
		}

		for _, c := range e.Dir {
			typ := sp.parseEntry(c)
			node.AddField(Field{
				name:     c.Name,
				value:    sp.parseEntry(c),
				nullable: c.Mandatory.Value(),
			})

			if c.Name == e.Key {
				node.SetKeyType(typ)
			}
		}
		return node
	}

	return Primitive{t: TS_ANY}
}

// ParseScheme produces a single interface from a
// yang module containing yang containers only
func (sp *Parser) ParseSchema() (Type, error, bool) {
	if sp.yang == nil {
		return nil, errors.New("Yang module is nil"), false
	}

	entry := yang.ToEntry(sp.yang)
	if entry == nil {
		return nil, errors.New("Yang entry is nil"), false
	}

	iface := Interface{
		name: entry.Name,
	}

	for _, entry := range entry.Dir {
		if entry.RPC != nil {
			// Ignored: in rpc statements
			continue
		}

		if entry.Config.Value() {
			// Ignored: used in action statements
			continue
		}

		iface.AddField(Field{
			name:     entry.Name,
			value:    sp.parseEntry(entry),
			nullable: entry.Mandatory.Value(),
		})
	}

	return iface, nil, iface.FieldCount() == 0
}

func (sp *Parser) defaultType() Type {
	if sp.strict {
		return Primitive{t: TS_UNKNOWN}
	} else {
		return Primitive{t: TS_ANY}
	}
}

func (sp *Parser) parseYangType(yType *yang.YangType) Type {
	if yType == nil {
		// TODO: panic?
		return sp.defaultType()
	}

	switch yType.Kind {
	case yang.Yint8:
		return Primitive{t: TS_NUMBER}
	case yang.Yint16:
		return Primitive{t: TS_NUMBER}
	case yang.Yint32:
		return Primitive{t: TS_NUMBER}
	case yang.Yint64:
		return Primitive{t: TS_BIGINT}
	case yang.Yuint8:
		return Primitive{t: TS_NUMBER}
	case yang.Yuint16:
		return Primitive{t: TS_NUMBER}
	case yang.Yuint32:
		return Primitive{t: TS_NUMBER}
	case yang.Yuint64:
		return Primitive{t: TS_NUMBER}
	case yang.Ydecimal64:
		return Primitive{t: TS_NUMBER}

	case yang.Ystring:
		return Primitive{t: TS_STRING}
	case yang.Ybinary:
		return Primitive{t: TS_STRING}
	case yang.Yidentityref:
		return Primitive{t: TS_STRING}
	case yang.YinstanceIdentifier:
		return Primitive{t: TS_STRING}

	case yang.Ybool:
		return Primitive{t: TS_BOOL}
	case yang.Yempty:
		// YANG 'empty' type is often represented as a boolean flag in TypeScript
		// where its presence indicates 'true'.
		return Primitive{t: TS_BOOL}

	case yang.Ybits:
		// YANG 'bits' type is typically represented as an array of strings in TypeScript,
		// where each string is a set bit.
		return Array{t: Primitive{t: TS_STRING}}

	case yang.Yleafref:
		// TODO: Leafref references another leaf. The actual type would be the type of the referenced leaf.
		// For simplicity, we'll default to string or any, as resolving the exact type
		// would require a more complex schema traversal and type resolution logic.
		return sp.defaultType()

	case yang.Ynone:
		return Primitive{t: TS_NEVER}

	case yang.Yenum:
		// For enumerations, generate a string literal union type for type safety.
		// Example: 'option-a' | 'option-b'
		var enumValues []string
		for _, value := range yType.Enum.ValueMap() {
			enumValues = append(enumValues, fmt.Sprintf("'%s'", value))
		}
		return Union{opts: unique(enumValues)}

	case yang.Yunion:
		var unionMembers []string

		for _, memberType := range yType.Type {
			t := sp.parseYangType(memberType)
			unionMembers = append(unionMembers, t.String())
		}

		return Union{opts: unique(unionMembers)}
	}

	return sp.defaultType()
}

func unique[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	result := []T{}
	for _, item := range slice {
		if _, found := seen[item]; !found {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
