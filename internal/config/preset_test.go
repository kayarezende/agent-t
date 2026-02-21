package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPresetSummary_Single(t *testing.T) {
	p := Preset{
		Name:    "my-preset",
		Project: "api",
		Layout:  "3,3",
		Tool:    "Claude Code",
	}
	got := p.Summary()
	want := "api | 3,3 | Claude Code"
	if got != want {
		t.Errorf("Summary() = %q, want %q", got, want)
	}
}

func TestPresetSummary_SingleNoTool(t *testing.T) {
	p := Preset{
		Name:    "my-preset",
		Project: "api",
		Layout:  "3,3",
		Tool:    "",
	}
	got := p.Summary()
	if !strings.Contains(got, "None") {
		t.Errorf("Summary() = %q, should contain 'None' when tool is empty", got)
	}
}

func TestPresetSummary_Split(t *testing.T) {
	p := Preset{
		Name:          "split-preset",
		Project:       "api",
		ProjectBottom: "frontend",
		Layout:        "3,3",
		Tool:          "Claude Code",
		ToolBottom:    "Codex",
	}
	got := p.Summary()
	// Should show both projects and both tools
	if !strings.Contains(got, "api") {
		t.Errorf("Summary() = %q, missing top project", got)
	}
	if !strings.Contains(got, "frontend") {
		t.Errorf("Summary() = %q, missing bottom project", got)
	}
	if !strings.Contains(got, "Claude Code") {
		t.Errorf("Summary() = %q, missing top tool", got)
	}
	if !strings.Contains(got, "Codex") {
		t.Errorf("Summary() = %q, missing bottom tool", got)
	}
	if !strings.Contains(got, "+") {
		t.Errorf("Summary() = %q, missing '+' separator", got)
	}
}

func TestPresetSummary_SplitNoBottomTool(t *testing.T) {
	p := Preset{
		Name:          "split-preset",
		Project:       "api",
		ProjectBottom: "frontend",
		Layout:        "3,3",
		Tool:          "Claude Code",
		ToolBottom:    "",
	}
	got := p.Summary()
	if !strings.Contains(got, "None") {
		t.Errorf("Summary() = %q, should show 'None' for empty bottom tool", got)
	}
}

func TestPresetYAML_OmitEmpty(t *testing.T) {
	p := Preset{
		Name:    "single",
		Project: "api",
		Layout:  "3,3",
		Tool:    "Claude Code",
	}
	data, err := yaml.Marshal(p)
	if err != nil {
		t.Fatalf("yaml.Marshal: %v", err)
	}
	s := string(data)
	if strings.Contains(s, "project_bottom") {
		t.Errorf("YAML should omit project_bottom when empty, got:\n%s", s)
	}
	if strings.Contains(s, "tool_bottom") {
		t.Errorf("YAML should omit tool_bottom when empty, got:\n%s", s)
	}
}

func TestPresetYAML_Split(t *testing.T) {
	p := Preset{
		Name:          "split",
		Project:       "api",
		ProjectBottom: "frontend",
		Layout:        "3,3",
		Tool:          "Claude Code",
		ToolBottom:    "Codex",
	}
	data, err := yaml.Marshal(p)
	if err != nil {
		t.Fatalf("yaml.Marshal: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, "project_bottom: frontend") {
		t.Errorf("YAML should include project_bottom, got:\n%s", s)
	}
	if !strings.Contains(s, "tool_bottom: Codex") {
		t.Errorf("YAML should include tool_bottom, got:\n%s", s)
	}
}

func TestPresetYAML_Roundtrip(t *testing.T) {
	original := Preset{
		Name:          "split",
		Project:       "api",
		ProjectBottom: "frontend",
		Layout:        "3,3",
		Tool:          "Claude Code",
		ToolBottom:    "Codex",
	}
	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("yaml.Marshal: %v", err)
	}
	var decoded Preset
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}
	if decoded.ProjectBottom != original.ProjectBottom {
		t.Errorf("ProjectBottom: got %q, want %q", decoded.ProjectBottom, original.ProjectBottom)
	}
	if decoded.ToolBottom != original.ToolBottom {
		t.Errorf("ToolBottom: got %q, want %q", decoded.ToolBottom, original.ToolBottom)
	}
}

func TestPresetYAML_BackwardCompatible(t *testing.T) {
	// Old YAML without split fields should deserialize fine
	oldYAML := `name: old-preset
project: api
layout: "3,3"
tool: Claude Code
`
	var p Preset
	if err := yaml.Unmarshal([]byte(oldYAML), &p); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}
	if p.ProjectBottom != "" {
		t.Errorf("ProjectBottom should be empty for old presets, got %q", p.ProjectBottom)
	}
	if p.ToolBottom != "" {
		t.Errorf("ToolBottom should be empty for old presets, got %q", p.ToolBottom)
	}
	if p.Name != "old-preset" {
		t.Errorf("Name: got %q, want %q", p.Name, "old-preset")
	}
}
