package config

import "fmt"

type Preset struct {
	Name          string `yaml:"name"`
	Project       string `yaml:"project"`
	Layout        string `yaml:"layout"`
	Tool          string `yaml:"tool"`
	ProjectBottom string `yaml:"project_bottom,omitempty"`
	ToolBottom    string `yaml:"tool_bottom,omitempty"`
}

func (p Preset) Summary() string {
	tool := p.Tool
	if tool == "" {
		tool = "None"
	}
	if p.ProjectBottom != "" {
		toolBottom := p.ToolBottom
		if toolBottom == "" {
			toolBottom = "None"
		}
		return fmt.Sprintf("%s + %s | %s | %s + %s", p.Project, p.ProjectBottom, p.Layout, tool, toolBottom)
	}
	return fmt.Sprintf("%s | %s | %s", p.Project, p.Layout, tool)
}
