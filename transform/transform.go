package transform

import (
	"fmt"
	"log"
	"strings"
	"yangts/casing"

	"github.com/openconfig/goyang/pkg/yang"
)

func inspect(e *yang.Entry) {
	fmt.Println()
	fmt.Println("// Name: ", e.Name)
	fmt.Println("// Kind: ", e.Kind.String())
	fmt.Println("// NodeKind: ", e.Node.Kind())
	fmt.Println("// Namespace.Name: ", e.Namespace().Name)
	fmt.Println("// Namespace.Description: ", e.Namespace().Description)
	fmt.Println()
}

func TypeScriptFromYangEntry(e *yang.Entry) (string, error) {
	if e == nil {
		panic("Entry is nil")
	}

	if e.Node.Kind() != "module" {
		panic("entry must be a module")
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("export type %s = %s;\n", casing.ToPascalCase(e.Name), parseContainer(e, 1)))

	return sb.String(), nil
}

func optional(e *yang.Entry) string {
	if e.Mandatory == yang.TSUnset {
		// Unset is required
		return ""
	}

	if e.Mandatory.Value() == true {
		return ""
	}

	return "?"
}

func parseContainer(e *yang.Entry, level int) string {
	if !e.IsContainer() {
		return fmt.Sprintf("// Skip: %s. Not a container", e.Name)
	}

	indent := strings.Repeat(" ", level*2)

	var sb strings.Builder

	sb.WriteString("{\n")

	for _, child := range e.Dir {

		if child.Node.Kind() == "rpc" {
			continue
		}

		if child.IsContainer() {
			ident := child.Name
			body := parseContainer(child, level+1)
			sb.WriteString(fmt.Sprintf("%s\"%s\"%s: %s;\n", indent, ident, optional(child), body))
		} else {
			if child.Type == nil {
				inspect(child)
				if child.Node.Kind() == "list" {
					tsType, err := yangTypeToTypeScriptType(child.Type)
					if err != nil {
						log.Fatal(err)
					}

					// TODO: we can determine the key type which is assoicated with the following record type
					sb.WriteString(fmt.Sprintf("%s\"%s\"%s: Record<string, %s>;\n", indent, child.Name, optional(child), tsType))
					continue
				}

				if child.Node.Kind() == "leaf-list" {
					tsType, err := yangTypeToTypeScriptType(child.Type)
					if err != nil {
						log.Fatal(err)
					}

					// TODO: we can determine the key type which is assoicated with the following record type
					sb.WriteString(fmt.Sprintf("%s\"%s\"%s: Array<%s>;\n", indent, child.Name, optional(child), tsType))
					continue
				}
			}

			tsType, err := yangTypeToTypeScriptType(child.Type)
			if err != nil {
				log.Fatal(err)
			}
			sb.WriteString(fmt.Sprintf("%s\"%s\"%s: %s;\n", indent, child.Name, optional(child), tsType))
		}
	}

	sb.WriteString(fmt.Sprintf("%s}", strings.Repeat(" ", (level-1)*2)))

	return sb.String()
}
