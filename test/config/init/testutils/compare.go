package testutils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/google/go-cmp/cmp"
)

func dirContent(dirPath string, excludeDirs ...string) (map[string]string, error) {
	content := make(map[string]string)

	excluded := make(map[string]bool)
	for _, dir := range excludeDirs {
		excluded[dir] = true
	}

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// If the directory is in the excluded list, skip it
		if d.IsDir() && excluded[d.Name()] {
			return fs.SkipDir
		}

		if !d.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			// Get relative path to maintain structure
			relativePath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}
			content[relativePath] = string(data)
		}
		return nil
	})

	return content, err
}

func CompareDirs(dir1, dir2 string, excludeDirs ...string) (bool, error) {
	content1, err := dirContent(dir1, excludeDirs...)
	if err != nil {
		return false, err
	}

	content2, err := dirContent(dir2, excludeDirs...)
	if err != nil {
		return false, err
	}

	if diff := cmp.Diff(content1, content2); diff != "" {
		fmt.Println(diff)
		return false, nil
	}

	return true, nil
}
