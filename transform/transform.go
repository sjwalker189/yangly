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

func docblock(e *yang.Entry, tabsize int) string {
	if e == nil {
		return ""
	}

	indent := strings.Repeat(" ", tabsize)

	var blocks []string

	if e.Description != "" {
		blocks = append(blocks, fmt.Sprintf("%s * %s", indent, e.Description))
	}

	if e.IsLeafList() {
		defaults := e.DefaultValues()
		if len(defaults) > 0 {
			blocks = append(blocks, fmt.Sprintf("%s * @default [%s]", indent, strings.Join(e.DefaultValues(), ", ")))
		}
	} else {
		val, ok := e.SingleDefaultValue()
		if ok {
			blocks = append(blocks, fmt.Sprintf("%s * @default %s", indent, val))
		}

	}

	if len(blocks) > 0 {
		return fmt.Sprintf("%s/**\n%s\n%s */\n", indent, strings.Join(blocks, "\n"), indent)
	}

	return ""
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

		if child.RPC != nil {
			continue
		}

		if child.IsContainer() {
			ident := child.Name
			body := parseContainer(child, level+1)
			sb.WriteString(fmt.Sprintf("%s\"%s\"%s: %s;\n", indent, ident, optional(child), body))
		} else {
			if child.Type == nil {
				if child.Kind == yang.DirectoryEntry && child.ListAttr != nil {

					keyType := "string"
					members := []string{}

					for _, leaf := range child.Dir {
						if leaf.IsContainer() {
							ms := fmt.Sprintf("%s\"%s\"%s: %s;", strings.Repeat(" ", (level+1)*2), leaf.Name, optional(leaf), parseContainer(leaf, level+1))
							members = append(members, docblock(leaf, (level+1)*2))
							members = append(members, ms)
						} else {
							tsType, err := yangTypeToTypeScriptType(leaf.Type)
							if err != nil {
								log.Fatal(err)
							}

							ms := fmt.Sprintf("%s\"%s\"%s: %s;", strings.Repeat(" ", (level+1)*2), leaf.Name, optional(leaf), tsType)
							members = append(members, docblock(leaf, (level+1)*2))
							members = append(members, ms)
							if leaf.Name == child.Key {
								keyType = tsType
							}
						}
					}

					if child.Type == nil {
						fmt.Printf("%+v\n\n", child.Dir)
					}

					sb.WriteString(docblock(child, level*2))
					sb.WriteString(fmt.Sprintf("%s\"%s\"%s: Record<%s, {\n%s\n%s}>;\n", indent, child.Name, optional(child), keyType, strings.Join(members, "\n"), indent))
					continue
				}

				if child.Node.Kind() == "leaf-list" {
					tsType, err := yangTypeToTypeScriptType(child.Type)
					if err != nil {
						log.Fatal(err)
					}

					sb.WriteString(docblock(child, level*2))
					sb.WriteString(fmt.Sprintf("%s\"%s\"%s: Array<%s>;\n", indent, child.Name, optional(child), tsType))
					continue
				}
			}

			tsType, err := yangTypeToTypeScriptType(child.Type)
			if err != nil {
				log.Fatal(err)
			}

			sb.WriteString(docblock(child, level*2))
			sb.WriteString(fmt.Sprintf("%s\"%s\"%s: %s;\n", indent, child.Name, optional(child), tsType))
		}
	}

	sb.WriteString(fmt.Sprintf("%s}", strings.Repeat(" ", (level-1)*2)))

	return sb.String()
}
