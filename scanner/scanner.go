package scanner

import (
	"fmt"
	"os"
	"path/filepath"
)

func ScanDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var files []string

	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(dir, name)
		if entry.IsDir() {
			f, err := ScanDir(path)
			if err != nil {
				return nil, err
			}
			files = append(files, f...)
		} else if filepath.Ext(name) == ".yang" {
			files = append(files, path)
		}
	}

	return files, nil
}

