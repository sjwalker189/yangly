package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/openconfig/goyang/pkg/yang"
)

func main() {

	yangDirs := []string{
		"./yangs/common",
		"./yangs/system",
	}

	yangFiles := []string{
		"./yangs/common/ietf-inet-types.yang",
		"./yangs/system/avnm-system-configuration.yang",
	}

	if len(yangFiles) == 0 {
		log.Println("No .yang files found in the specified directory.")
		return
	}

	// Create a new Modules instance
	moduleSet := yang.NewModules()
	for _, path := range yangDirs {
		moduleSet.AddPath(path)
	}

	// Read each YANG file
	for _, file := range yangFiles {
		// log.Printf("Reading YANG file: %s", file)
		if err := moduleSet.Read(file); err != nil {
			log.Printf("Error reading %s: %v", file, err)
		}
	}

	// Process the modules (resolves imports, includes, etc.)
	if errs := moduleSet.Process(); errs != nil {
		for _, e := range errs {
			log.Printf("Error processing modules: %v", e)
		}
		log.Fatalf("Failed to process YANG modules.")
	}

	// Access the parsed modules as yang.Entry trees
	// You can iterate through moduleSet.Modules to get the top-level modules
	// and then convert them to yang.Entry trees.
	for name, module := range moduleSet.Modules {
		// fmt.Printf("  Module: %s (Namespace: %s)\n", name, module.Namespace.Name)
		entry := yang.ToEntry(module)
		if entry != nil {
			// fmt.Printf("    Root entry for %s: %s\n", name, entry.Path())
			TraverseAndGenerateTS(entry, 0)
		} else {
			log.Fatalf("    Could not convert module %s to Entry tree.\n", name)
		}
	}
}

// getYangFiles walks a directory and returns a list of all .yang files.
func getYangFiles(dir string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yang" {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files, nil
}

// traverseAndGenerateTS recursively inspects a yang.Entry node and generates TypeScript
// declarations for leaf, leaf-list, container, and list nodes it encounters.
func TraverseAndGenerateTS(e *yang.Entry, indentLevel int) {
	// Generate declarations based on the node type
	if e.IsLeaf() || e.IsLeafList() {
		tsDecl, err := GenerateTypeScriptLeafType(e) // Renamed function for clarity
		if err != nil {
			log.Printf("Error generating TypeScript for %s (%s): %v", e.Name, e.Path(), err)
		} else {
			fmt.Println(tsDecl)
		}
	} else if e.IsContainer() || e.IsList() {
		tsDecl, err := GenerateTypeScriptInterface(e, indentLevel, true) // New function for containers/lists
		if err != nil {
			log.Printf("Error generating TypeScript for %s (%s): %v", e.Name, e.Path(), err)
		} else {
			fmt.Println(tsDecl)
		}
	}

	// Recursively inspect child nodes (containers, lists, choices, cases, leaves, leaf-lists)
	// We iterate through all children here, but the generation functions above filter
	// for specific types. This ensures all parts of the tree are visited.
	for _, child := range e.Dir {
		TraverseAndGenerateTS(child, indentLevel+1)
	}
}

// yangTypeToTypeScriptType converts a yang.YangType (which contains Kind and other details)
// to its corresponding TypeScript type string. It handles complex types like enums and unions.
func yangTypeToTypeScriptType(yType *yang.YangType) (string, error) {
	if yType == nil {
		return "", fmt.Errorf("yang type is nil")
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

// toPascalCase converts a kebab-case or snake_case string to PascalCase.
// E.g., "my-string-leaf" -> "MyStringLeaf"
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_'
	})
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = string(unicode.ToUpper(rune(part[0]))) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// toCamelCase converts a kebab-case or snake_case string to camelCase.
// E.g., "my-string-leaf" -> "myStringLeaf"
func toCamelCase(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_'
	})
	if len(parts) == 0 {
		return ""
	}
	// First part is lowercase
	result := []rune(strings.ToLower(parts[0]))
	// Subsequent parts are PascalCase
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result = append(result, unicode.ToUpper(rune(parts[i][0])))
			result = append(result, []rune(parts[i][1:])...)
		}
	}
	return string(result)
}

// GenerateTypeScriptLeafType generates a TypeScript type declaration string for a given yang.Entry.
// It specifically handles leaf and leaf-list nodes, returning an error if the entry is not one of these.
func GenerateTypeScriptLeafType(e *yang.Entry) (string, error) {
	if e == nil {
		return "", fmt.Errorf("yang.Entry is nil")
	}

	// Ensure the entry is a leaf or leaf-list
	if !e.IsLeaf() && !e.IsLeafList() {
		return "", fmt.Errorf("entry '%s' is not a leaf or leaf-list node (Kind: %s)", e.Name, e.Kind)
	}

	// A leaf or leaf-list must have a type
	if e.Type == nil {
		return "", fmt.Errorf("leaf/leaf-list entry '%s' has no type defined", e.Name)
	}

	// Generate a PascalCase type name for the TypeScript declaration.
	// We append "Type" to avoid potential naming conflicts with data properties.
	tsTypeName := toPascalCase(e.Name) + "Type"

	// Get the TypeScript equivalent of the YANG base type
	tsBaseType, err := yangTypeToTypeScriptType(e.Type)
	if err != nil {
		return "", fmt.Errorf("failed to convert YANG type for '%s': %w", e.Name, err)
	}

	var declaration strings.Builder
	// Add JSDoc comments for the YANG path and description
	declaration.WriteString(fmt.Sprintf("/**\n * YANG Path: /%s\n", e.Path())) // e.Path() gives the full schema path
	if e.Description != "" {
		// Indent multiline descriptions for better JSDoc formatting
		descriptionLines := strings.Split(e.Description, "\n")
		for _, line := range descriptionLines {
			declaration.WriteString(fmt.Sprintf(" * %s\n", strings.TrimSpace(line)))
		}
	}
	declaration.WriteString(" */\n")

	// Determine if it's a leaf-list and adjust the TypeScript type accordingly
	if e.IsLeafList() {
		declaration.WriteString(fmt.Sprintf("export type %s = %s[];\n", tsTypeName, tsBaseType))
	} else { // It's a leaf
		declaration.WriteString(fmt.Sprintf("export type %s = %s;\n", tsTypeName, tsBaseType))
	}

	return declaration.String(), nil
}

// GenerateTypeScriptInterface generates a TypeScript interface string for a given yang.Entry
// representing a container or a list entry.
func GenerateTypeScriptInterface(e *yang.Entry, indentLevel int, toplevel bool) (string, error) {
	if e == nil {
		return "", fmt.Errorf("yang.Entry is nil")
	}

	if !e.IsContainer() && !e.IsList() {
		return "", fmt.Errorf("entry '%s' is not a container or list node (Kind: %s)", e.Name, e.Kind)
	}

	// Use PascalCase for interface names
	tsInterfaceName := toPascalCase(e.Name)
	var declaration strings.Builder
	indent := strings.Repeat("  ", indentLevel)

	// Add JSDoc comments
	declaration.WriteString(fmt.Sprintf("%s/**\n", indent))
	declaration.WriteString(fmt.Sprintf("%s * YANG Path: /%s\n", indent, e.Path()))
	if e.Description != "" {
		descriptionLines := strings.Split(e.Description, "\n")
		for _, line := range descriptionLines {
			declaration.WriteString(fmt.Sprintf("%s * %s\n", indent, strings.TrimSpace(line)))
		}
	}

	if toplevel {
		declaration.WriteString(fmt.Sprintf("%s */\n", indent))
		declaration.WriteString(fmt.Sprintf("%sexport interface %s {\n", indent, tsInterfaceName))
	} else {
		declaration.WriteString(fmt.Sprintf("%s%s {\n", indent, tsInterfaceName))
	}

	// Iterate through children to add properties
	// Sort children by name for deterministic output
	var childNames []string
	for name := range e.Dir {
		childNames = append(childNames, name)
	}
	sort.Strings(childNames)

	for _, childName := range childNames {
		child := e.Dir[childName]
		// Only process data nodes (leaf, leaf-list, container, list) that are direct children
		// Skip choices and cases themselves, their children will be processed if they are data nodes.
		if child.Kind == yang.LeafEntry || child.Kind == yang.DirectoryEntry {
			propName := fmt.Sprintf("\"%s\"", child.Name)
			propType := ""
			var err error

			if child.IsLeaf() || child.IsLeafList() {
				propType, err = yangTypeToTypeScriptType(child.Type)
				if err != nil {
					log.Printf("Warning: Could not determine type for leaf/leaf-list %s: %v", child.Path(), err)
					propType = "any" // Fallback
				}
				if child.IsLeafList() {
					propType += "[]"
				}
			} else if child.IsContainer() || child.IsList() {
				// For nested containers/lists, reference their interface name
				propType = toPascalCase(child.Name)
				if child.IsList() {
					propType += "[]"
				}
			} else {
				// This case should ideally not be hit due to the outer if, but as a safeguard
				log.Printf("Warning: Unhandled child kind for property %s: %s", child.Path(), child.Kind)
				propType = "any"
			}

			// Add property to interface (making all properties optional for simplicity)
			declaration.WriteString(fmt.Sprintf("%s  %s%s: %s;\n", indent, strings.Repeat("  ", 1), propName, propType))
		}
	}

	declaration.WriteString(fmt.Sprintf("%s}\n", indent))

	// If it's a list, also generate a type alias for the array of the interface
	if e.IsList() {
		declaration.WriteString(fmt.Sprintf("\n%s/**\n", indent))
		declaration.WriteString(fmt.Sprintf("%s * YANG Path: /%s\n", indent, e.Path()))
		if e.Description != "" {
			descriptionLines := strings.Split(e.Description, "\n")
			for _, line := range descriptionLines {
				declaration.WriteString(fmt.Sprintf("%s *  %s\n", indent, strings.TrimSpace(line)))
			}
		}
		declaration.WriteString(fmt.Sprintf("%s */\n", indent))
		declaration.WriteString(fmt.Sprintf("%sexport type %sListType = %s[];\n", indent, tsInterfaceName, tsInterfaceName))
	}

	return declaration.String(), nil
}
