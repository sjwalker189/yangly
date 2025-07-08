package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"yangly/ast"
	"yangly/scanner"

	"github.com/sjwalker189/goyang/pkg/yang"
	"github.com/spf13/cobra"
)

var scandir string
var outdir string

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate TypeScript interfaces from YANG schemas.",
	Run: func(cmd *cobra.Command, args []string) {
		// Discover all .yang files
		files, err := scanner.ScanDir(scandir)

		if err != nil {
			log.Fatalf("Failed to scan directory %s\n", err)
		}

		if len(files) == 0 {
			log.Fatal("No .yang files found")
		}

		moduleSet := yang.NewModules()

		// Read all yangs
		for _, file := range files {
			if err := moduleSet.Read(file); err != nil {
				log.Fatal("Failed to read yang module. ", err)
			}
		}

		// Process the yang modules to build the ast
		if errs := moduleSet.Process(); errs != nil {
			for _, err := range errs {
				log.Fatal("Error processing modules: ", err)
			}
			log.Fatal("Failed to process YANG modules")
		}

		// Transformations

		for name, module := range moduleSet.Modules {
			// Revisions cause duplicate modules to be created. We only care about
			// the latest revision.
			if module.Current() != "" {
				if !strings.HasSuffix(name, module.Current()) {
					continue
				}
			}

			log.Println("Process module: ", module.Name)

			parser := ast.NewParser(module)
			typ, err, empty := parser.ParseSchema()
			if err != nil {
				panic(err)
			}

			if empty {
				log.Println("No schemas found in: ", module.Name)
				continue
			}

			fpath := filepath.Join(fmt.Sprintf("types/%s.ts", module.Name))
			file, err := os.Create(fpath)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			file.WriteString(typ.String())
		}
	},
}

func init() {
	genCmd.Flags().StringVarP(&scandir, "path", "p", "yangs", "Directory to scan for .yang files.")
	genCmd.Flags().StringVarP(&outdir, "out", "o", "dist", "Directory to outpult .ts files.")
	rootCmd.AddCommand(genCmd)
}
