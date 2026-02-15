package scanner

import (
	"os"
	"path/filepath"
	"sort"
)

type Project struct {
	Name string
	Path string
}

func Scan(baseDir string) ([]Project, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}

	var projects []Project
	for _, e := range entries {
		if !e.IsDir() || e.Name()[0] == '.' {
			continue
		}
		projects = append(projects, Project{
			Name: e.Name(),
			Path: filepath.Join(baseDir, e.Name()),
		})
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}
