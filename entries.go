package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Entries struct {
	Path  string
	Files []string
}

var supportedExtensions = regexp.MustCompile(`.jpg|.jpeg|.png|.webp`)

func search(root string, maxDepth int) []Entries {
	result := make(map[string]*Entries)
	baseDepth := getDepth(root)

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if getDepth(path)-1 > baseDepth+maxDepth {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		dir := filepath.Dir(path)
		entries, ok := result[dir]
		if !ok {
			e := &Entries{
				Path:  dir,
				Files: []string{},
			}
			result[dir] = e
			entries = e
		}

		if isSupportedExtension(path) {
			entries.Files = append(entries.Files, path)
		}

		return nil
	})

	entries := make([]Entries, 0, len(result))

	for _, entry := range result {
		if len(entry.Files) > 0 {
			entries = append(entries, *entry)
		}
	}

	return entries
}

func getDepth(path string) int {
	return strings.Count(path, string(os.PathSeparator))
}

func isSupportedExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return supportedExtensions.MatchString(ext)
}
