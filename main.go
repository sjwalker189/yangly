package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"yangts/transform"

	"github.com/openconfig/goyang/pkg/yang"
)

func main() {
	yangFiles := []string{
		"./yangs/example.yang",
	}

	if len(yangFiles) == 0 {
		log.Println("No .yang files found in the specified directory.")
		return
	}

	moduleSet := yang.NewModules()

	for _, file := range yangFiles {
		log.Printf("Reading YANG file: %s", file)
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

	log.Printf("Processing %d modules\n", len(moduleSet.Modules))

	// Access the parsed modules as yang.Entry trees
	// You can iterate through moduleSet.Modules to get the top-level modules
	// and then convert them to yang.Entry trees.

	// IMPORTANT: Revisions cause duplicate modules

	for name, module := range moduleSet.Modules {
		entry := yang.ToEntry(module)
		if entry != nil {
			// TODO: Provide bufio writer into transformer
			filename := fmt.Sprintf("%s.ts", module.Name)
			header := fmt.Sprintf("// Module: %s (Namespace: %s)\n", name, module.Namespace.Name)
			contents, err := transform.TypeScriptFromYangEntry(entry)

			writeFile(filename, contents)

			if err != nil {
				panic(err)
			}

			fmt.Printf("// File: %s\n%s\n%s\n", filename, header, contents)
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

func writeFile(filename string, content string) {
	file, err := os.Create(filepath.Join("./out", filename))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.WriteString(content)
}
