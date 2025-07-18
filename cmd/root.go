package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"yangly/ast"
	"yangly/scanner"

	"github.com/sjwalker189/goyang/pkg/yang"
	"github.com/spf13/cobra"
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&scandir, "path", "p", "yangs", "Directory to scan for .yang files.")
	rootCmd.Flags().StringVarP(&outdir, "out", "o", "dist", "Directory to outpult .ts files.")
	rootCmd.Flags().BoolVarP(&bail, "bail", "b", false, "Exit the process an error is encountered.")
}

var scandir string
var outdir string
var bail bool

var errorCount int

func infoln(v ...any) {
	fmt.Println(v...)
}

func errorln(v ...any) {
	if bail {
		fmt.Println(v...)
		os.Exit(1)
	} else {
		fmt.Println(v...)
		errorCount += 1
	}
}

var rootCmd = &cobra.Command{
	Use:   "yangly",
	Short: "Generate TypeScript interfaces from YANG schemas.",
	Run: func(cmd *cobra.Command, args []string) {
		var p string
		if filepath.IsAbs(scandir) {
			p = scandir
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				errorln("Failed to scan directory: ", scandir)
			}
			p = filepath.Join(cwd, scandir)
		}

		infoln("Scanning directory: ", p)

		// Discover all .yang files
		files, err := scanner.ScanDir(p)
		if err != nil {
			infoln("Failed to scan directory: ", p)
			os.Exit(0)
		}

		if len(files) == 0 {
			infoln("No .yang files found")
			os.Exit(0)
		}

		moduleSet := yang.NewModules()

		// Read all yangs
		infoln("Reading files: ")
		infoln(strings.Join(files, "\n"))

		for _, file := range files {
			moduleSet.AddPath(filepath.Dir(file))
			if err := moduleSet.Read(file); err != nil {
				errorln("Failed to read yang module. ", err)
			}
		}

		// Process the yang modules to build the ast
		if errs := moduleSet.Process(); errs != nil {
			for _, err := range errs {
				errorln("Error processing modules: ", err)
			}
			errorln("Failed to process YANG modules")
		}

		// Transformations

		infoln("Preparing output directory:", outdir)
		if err := os.MkdirAll(outdir, os.ModePerm); err != nil {
			errorln(err)
		}

		for name, module := range moduleSet.Modules {
			// Revisions cause duplicate modules to be created. We only care about
			// the latest revision.
			if module.Current() != "" {
				if !strings.HasSuffix(name, module.Current()) {
					continue
				}
			}

			parser := ast.NewParser(module)
			typ, err, empty := parser.ParseSchema()
			if err != nil {
				errorln(err)
			}

			if empty {
				// TODO: verbose
				// infoln("No schemas found in: ", module.Name, "Skipping...")
				continue
			}

			infoln("Processed module: ", module.Name)

			fpath := filepath.Join(outdir, fmt.Sprintf("%s.ts", module.Name))
			infoln("Writing typescript output: ", fpath)
			file, err := os.Create(fpath)
			if err != nil {
				errorln(err)
			}
			file.WriteString(typ.String())
			file.Close()
		}

		infoln("Completed with", errorCount, "errors.")
	},
}
