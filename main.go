package main

import (
	"fmt"
	"github.com/sjwalker189/goyang/pkg/yang"
	"log"
	"os"
	"path/filepath"
	"strings"
	"yangts/transform"
)

func main() {
	var yangFiles []string

	files, _ := scanFiles("./derived-yangs")
	for _, fp := range files {
		yangFiles = append(yangFiles, fp)
	}

	if len(yangFiles) == 0 {
		log.Println("No .yang files found in the specified directory.")
		return
	}

	moduleSet := yang.NewModules()
	moduleSet.AddPath("./derived-yangs")

	for _, file := range yangFiles {
		fmt.Printf("Reading YANG file: %s\n", file)
		if err := moduleSet.Read(file); err != nil {
			fmt.Printf("Error reading %s: %v\n", file, err)
		}
	}

	// Process the modules (resolves imports, includes, etc.)
	if errs := moduleSet.Process(); errs != nil {
		for _, e := range errs {
			fmt.Println("Error processing modules: ", e)
		}
		fmt.Println("Failed to process YANG modules.")
	}

	fmt.Printf("Processing %d modules\n", len(moduleSet.Modules))

	// Access the parsed modules as yang.Entry trees
	// You can iterate through moduleSet.Modules to get the top-level modules
	// and then convert them to yang.Entry trees.

	for name, module := range moduleSet.Modules {
		if module.Current() != "" {
			// Revisions cause duplicate modules to be created. We only care about
			// the latest revision.
			if !strings.HasSuffix(name, module.Current()) {
				continue
			}
		}

		entry := yang.ToEntry(module)

		if entry != nil {
			filepath := fmt.Sprintf("dist/types/%s.ts", module.Name)
			contents, err := transform.TypeScriptFromYangEntry(entry)
			if err != nil {
				panic(err)
			}

			body := fmt.Sprintf("// Module: %s (Namespace: %s)\n\n%s\n", name, module.Namespace.Name, contents)
			writeFile(filepath, body)
		} else {
			fmt.Printf("    Could not convert module %s to Entry tree.\n", name)
		}
	}

}

func scanFiles(dir string) ([]string, error) {
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

func writeFile(filename string, content string) {
	file, err := os.Create(filepath.Join("./out", filename))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.WriteString(content)
}
