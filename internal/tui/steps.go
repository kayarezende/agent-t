package tui

import (
	"fmt"

	"agent-t/internal/config"
	"agent-t/internal/scanner"

	"github.com/charmbracelet/bubbles/list"
)

type step int

const (
	stepPreset  step = iota
	stepProject step = iota
	stepLayout  step = iota
	stepTool    step = iota
	stepConfirm step = iota
	stepDone    step = iota
)

func stepTitle(s step) string {
	switch s {
	case stepPreset:
		return "Choose a Preset"
	case stepProject:
		return "Select Project"
	case stepLayout:
		return "Select Layout"
	case stepTool:
		return "Select AI Tool"
	case stepConfirm:
		return "Confirm & Launch"
	default:
		return ""
	}
}

func stepNumber(s step, hasPresets bool) (int, int) {
	total := 4
	switch s {
	case stepPreset:
		return 0, total // presets shown as step 0 (not counted)
	case stepProject:
		return 1, total
	case stepLayout:
		return 2, total
	case stepTool:
		return 3, total
	case stepConfirm:
		return 4, total
	}
	return 0, total
}

// Layout represents a terminal grid layout.
type Layout struct {
	Name string
	Cols int
	Rows int
	Desc string
}

func (l Layout) ID() string { return fmt.Sprintf("%dx%d", l.Cols, l.Rows) }

var Layouts = []Layout{
	{Name: "2 terminals", Cols: 2, Rows: 1, Desc: "[ ][ ]"},
	{Name: "4 terminals", Cols: 2, Rows: 2, Desc: "[ ][ ] / [ ][ ]"},
	{Name: "6 terminals", Cols: 3, Rows: 2, Desc: "[ ][ ][ ] / [ ][ ][ ]"},
	{Name: "8 terminals", Cols: 4, Rows: 2, Desc: "[ ][ ][ ][ ] / [ ][ ][ ][ ]"},
}

// Tool represents an AI tool or command to run in each terminal.
type Tool struct {
	Name    string
	Command string
	Custom  bool
}

var BuiltinTools = []Tool{
	{Name: "None - just terminals", Command: ""},
	{Name: "Claude Code", Command: "claude"},
	{Name: "Claude Code (Chrome)", Command: "claude --chat-mode browser"},
	{Name: "Codex", Command: "codex"},
	{Name: "OpenCode", Command: "opencode"},
}

func AllTools(cfg *config.Config) []Tool {
	tools := make([]Tool, len(BuiltinTools))
	copy(tools, BuiltinTools)
	for name, cmd := range cfg.CustomCommands {
		tools = append(tools, Tool{Name: name, Command: cmd, Custom: true})
	}
	return tools
}

// --- List item types ---

type presetItem struct {
	preset config.Preset
	isNew  bool
}

func (i presetItem) Title() string {
	if i.isNew {
		return "New workspace..."
	}
	return i.preset.Name
}
func (i presetItem) Description() string {
	if i.isNew {
		return "Start fresh — pick project, layout, tool"
	}
	return i.preset.Summary()
}
func (i presetItem) FilterValue() string { return i.Title() }

type projectItem struct {
	project scanner.Project
}

func (i projectItem) Title() string       { return i.project.Name }
func (i projectItem) Description() string { return i.project.Path }
func (i projectItem) FilterValue() string { return i.project.Name }

type layoutItem struct {
	layout Layout
}

func (i layoutItem) Title() string       { return i.layout.Name }
func (i layoutItem) Description() string { return i.layout.Desc }
func (i layoutItem) FilterValue() string { return i.layout.Name }

type toolItem struct {
	tool Tool
}

func (i toolItem) Title() string {
	if i.tool.Custom {
		return i.tool.Name + " (custom)"
	}
	return i.tool.Name
}
func (i toolItem) Description() string {
	if i.tool.Command == "" {
		return "Just open terminals"
	}
	return i.tool.Command
}
func (i toolItem) FilterValue() string { return i.tool.Name }

type confirmItem struct {
	name string
	desc string
}

func (i confirmItem) Title() string       { return i.name }
func (i confirmItem) Description() string { return i.desc }
func (i confirmItem) FilterValue() string { return i.name }

// --- List builders ---

func newPresetList(presets []config.Preset, width, height int) list.Model {
	items := []list.Item{presetItem{isNew: true}}
	for _, p := range presets {
		items = append(items, presetItem{preset: p})
	}
	l := list.New(items, newStyledDelegate(), width, height)
	l.Title = "Presets"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(len(presets) > 5)
	l.SetShowHelp(true)
	// Put cursor on first preset (index 1), not "New workspace..."
	if len(presets) > 0 {
		l.Select(1)
	}
	return l
}

func newProjectList(projects []scanner.Project, width, height int) list.Model {
	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = projectItem{project: p}
	}
	l := list.New(items, newStyledDelegate(), width, height)
	l.Title = "Projects"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	return l
}

func newLayoutList(width, height int, defaultLayout string) list.Model {
	items := make([]list.Item, len(Layouts))
	defaultIdx := 0
	for i, lay := range Layouts {
		items[i] = layoutItem{layout: lay}
		if lay.ID() == defaultLayout {
			defaultIdx = i
		}
	}
	l := list.New(items, newStyledDelegate(), width, height)
	l.Title = "Layout"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.Select(defaultIdx)
	return l
}

func newToolList(tools []Tool, width, height int, defaultTool string) list.Model {
	items := make([]list.Item, len(tools))
	defaultIdx := 0
	for i, t := range tools {
		items[i] = toolItem{tool: t}
		if t.Name == defaultTool {
			defaultIdx = i
		}
	}
	l := list.New(items, newStyledDelegate(), width, height)
	l.Title = "AI Tool"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	l.Select(defaultIdx)
	return l
}

func newConfirmList(width, height int) list.Model {
	items := []list.Item{
		confirmItem{name: "Launch", desc: "Open terminals now"},
		confirmItem{name: "Save as preset & Launch", desc: "Save this combo for quick access next time"},
	}
	l := list.New(items, newStyledDelegate(), width, height)
	l.Title = "Ready?"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	return l
}
