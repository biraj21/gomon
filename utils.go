package main

import (
	"os"
	"path/filepath"
	"strings"
)

func getAllFiles(dir string, ext string) ([]string, error) {
	ext = strings.ToLower(strings.Trim(ext, " "))
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ext) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}
