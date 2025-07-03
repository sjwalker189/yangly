package transform

import (
	"fmt"
	"strings"

	"github.com/openconfig/goyang/pkg/yang"
)

func yangTypeToTypeScriptType(yType *yang.YangType) (string, error) {
	if yType == nil {
		// TODO: yType can be nil if the type is an identifier for a typedef
		return "any", nil
	}

	switch yType.Kind {

	// TODO: 64bit number types should use typescript BigInt
	case yang.Yint8, yang.Yint16, yang.Yint32, yang.Yint64, yang.Yuint8, yang.Yuint16, yang.Yuint32, yang.Yuint64, yang.Ydecimal64:
		return "number", nil
	case yang.Ystring, yang.Ybinary:
		return "string", nil
	case yang.Ybool:
		return "boolean", nil
	case yang.Yempty:
		// YANG 'empty' type is often represented as a boolean flag in TypeScript
		// where its presence indicates 'true'.
		return "boolean", nil
	case yang.Ybits:
		// YANG 'bits' type is typically represented as an array of strings in TypeScript,
		// where each string is a set bit.
		return "string[]", nil
	case yang.Yenum:
		// For enumerations, generate a string literal union type for type safety.
		// Example: 'option-a' | 'option-b'
		var enumValues []string
		for _, value := range yType.Enum.ValueMap() {
			enumValues = append(enumValues, fmt.Sprintf("'%s'", value))
		}
		if len(enumValues) == 0 {
			return "string", nil // Fallback if no enum values defined
		}
		return strings.Join(enumValues, " | "), nil
	case yang.Yidentityref:
		// TODO: Resolve the reference type
		// Identityref points to an identity, typically represented as a string in data.
		return "string", nil
	case yang.Yleafref:
		// TODO: Leafref references another leaf. The actual type would be the type of the referenced leaf.
		// For simplicity, we'll default to string or any, as resolving the exact type
		// would require a more complex schema traversal and type resolution logic.
		return "string", nil // Or "any" if you prefer less strictness
	case yang.Yunion:
		// For unions, recursively get the TypeScript type for each member and join them with '|'.
		var unionMembers []string
		for _, memberType := range yType.Type {
			tsType, err := yangTypeToTypeScriptType(memberType)
			if err != nil {
				return "", fmt.Errorf("failed to convert union member type %s: %w", memberType.Name, err)
			}
			unionMembers = append(unionMembers, tsType)
		}
		if len(unionMembers) == 0 {
			return "any", nil // Fallback for an empty union (shouldn't happen in valid YANG)
		}
		return strings.Join(unionMembers, " | "), nil
	case yang.YinstanceIdentifier:
		// YANG instance-identifier maps to a string path
		return "string", nil
	}

	// Catch-all for unhandled YANG base types
	return "any", fmt.Errorf("unhandled YANG base type: %s for leaf '%s'", yType.Kind, yType.Name)
}
