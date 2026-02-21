package tui

import (
	"strconv"
	"strings"

	"agent-t/internal/config"
	"agent-t/internal/scanner"

	"github.com/charmbracelet/bubbles/list"
)

type step int

const (
	stepPreset        step = iota
	stepMode          step = iota
	stepProject       step = iota
	stepProjectBottom step = iota
	stepLayout        step = iota
	stepTool          step = iota
	stepToolBottom    step = iota
	stepConfirm       step = iota
	stepDone          step = iota
)

func stepTitle(s step, splitMode bool) string {
	switch s {
	case stepPreset:
		return "Choose a Preset"
	case stepMode:
		return "Workspace Mode"
	case stepProject:
		if splitMode {
			return "Select Top Project"
		}
		return "Select Project"
	case stepProjectBottom:
		return "Select Bottom Project"
	case stepLayout:
		return "Select Layout"
	case stepTool:
		if splitMode {
			return "Select Top Tool"
		}
		return "Select AI Tool"
	case stepToolBottom:
		return "Select Bottom Tool"
	case stepConfirm:
		return "Confirm & Launch"
	default:
		return ""
	}
}

func stepNumber(s step, hasPresets bool, splitMode bool) (int, int) {
	if splitMode {
		total := 7
		switch s {
		case stepPreset:
			return 0, total
		case stepMode:
			return 1, total
		case stepProject:
			return 2, total
		case stepProjectBottom:
			return 3, total
		case stepLayout:
			return 4, total
		case stepTool:
			return 5, total
		case stepToolBottom:
			return 6, total
		case stepConfirm:
			return 7, total
		}
		return 0, total
	}

	total := 4
	switch s {
	case stepPreset:
		return 0, total
	case stepMode:
		return 1, total
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
	Name    string
	RowCols []int  // columns per row, e.g. [3,4] = 3 top, 4 bottom
	Desc    string
	Custom  bool
}

func (l Layout) TotalTerminals() int {
	n := 0
	for _, c := range l.RowCols {
		n += c
	}
	return n
}

func (l Layout) NumRows() int { return len(l.RowCols) }

func (l Layout) ID() string {
	parts := make([]string, len(l.RowCols))
	for i, c := range l.RowCols {
		parts[i] = strconv.Itoa(c)
	}
	return strings.Join(parts, ",")
}

func (l Layout) GenerateDesc() string {
	rows := make([]string, len(l.RowCols))
	for i, cols := range l.RowCols {
		cells := make([]string, cols)
		for j := range cells {
			cells[j] = "[ ]"
		}
		rows[i] = strings.Join(cells, "")
	}
	return strings.Join(rows, " / ")
}

var Layouts = []Layout{
	{Name: "2 terminals", RowCols: []int{2}, Desc: "[ ][ ]"},
	{Name: "3 terminals", RowCols: []int{3}, Desc: "[ ][ ][ ]"},
	{Name: "4 terminals (grid)", RowCols: []int{2, 2}, Desc: "[ ][ ] / [ ][ ]"},
	{Name: "4 terminals (vertical)", RowCols: []int{1, 1, 1, 1}, Desc: "[ ] / [ ] / [ ] / [ ]"},
	{Name: "6 terminals", RowCols: []int{3, 3}, Desc: "[ ][ ][ ] / [ ][ ][ ]"},
	{Name: "8 terminals", RowCols: []int{4, 4}, Desc: "[ ][ ][ ][ ] / [ ][ ][ ][ ]"},
}

// AllLayouts returns built-in layouts + custom layouts from config + a "Custom..." entry.
func AllLayouts(cfg *config.Config) []Layout {
	layouts := make([]Layout, len(Layouts))
	copy(layouts, Layouts)
	for _, cl := range cfg.CustomLayouts {
		l := Layout{Name: cl.Name, RowCols: cl.RowCols, Custom: true}
		l.Desc = l.GenerateDesc()
		layouts = append(layouts, l)
	}
	layouts = append(layouts, Layout{Name: "Custom...", Desc: "Define your own layout"})
	return layouts
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
	{Name: "Claude Code (Chrome)", Command: "claude --chrome"},
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

// SplitLayouts contains only multi-row layouts suitable for split workspace mode.
var SplitLayouts = []Layout{
	{Name: "4 terminals (2+2)", RowCols: []int{2, 2}, Desc: "[ ][ ] / [ ][ ]"},
	{Name: "6 terminals (3+3)", RowCols: []int{3, 3}, Desc: "[ ][ ][ ] / [ ][ ][ ]"},
	{Name: "8 terminals (4+4)", RowCols: []int{4, 4}, Desc: "[ ][ ][ ][ ] / [ ][ ][ ][ ]"},
}

// --- List item types ---

type modeItem struct {
	name string
	desc string
	split bool
}

func (i modeItem) Title() string       { return i.name }
func (i modeItem) Description() string { return i.desc }
func (i modeItem) FilterValue() string { return i.name }

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

func newModeList(width, height int) list.Model {
	items := []list.Item{
		modeItem{name: "Single Project", desc: "All terminals in one project", split: false},
		modeItem{name: "Split Workspace", desc: "Top row = project A, bottom row = project B", split: true},
	}
	l := list.New(items, newStyledDelegate(), width, height)
	l.Title = "Workspace Mode"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	return l
}

func newSplitLayoutList(width, height int, defaultLayout string) list.Model {
	items := make([]list.Item, len(SplitLayouts))
	defaultIdx := 0
	for i, lay := range SplitLayouts {
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

func newLayoutList(layouts []Layout, width, height int, defaultLayout string) list.Model {
	items := make([]list.Item, len(layouts))
	defaultIdx := 0
	for i, lay := range layouts {
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
