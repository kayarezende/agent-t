package config

import "fmt"

type Preset struct {
	Name    string `yaml:"name"`
	Project string `yaml:"project"`
	Layout  string `yaml:"layout"`
	Tool    string `yaml:"tool"`
}

func (p Preset) Summary() string {
	tool := p.Tool
	if tool == "" {
		tool = "None"
	}
	return fmt.Sprintf("%s | %s | %s", p.Project, p.Layout, tool)
}
