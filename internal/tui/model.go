package tui

import (
	"fmt"
	"strconv"
	"strings"

	"agent-t/internal/config"
	"agent-t/internal/scanner"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	currentStep step
	cancelled   bool
	configDirty bool

	width  int
	height int

	projects []scanner.Project
	cfg      *config.Config
	cwd      string
	tools    []Tool
	layouts  []Layout

	list list.Model

	selectedPreset  *config.Preset
	selectedProject scanner.Project
	selectedLayout  Layout
	selectedTool    Tool

	// Split workspace mode
	splitMode             bool
	selectedBottomProject scanner.Project
	selectedToolBottom    Tool

	// Preset naming
	namingPreset bool
	presetInput  textinput.Model

	// Custom layout input
	enteringCustomLayout bool
	customLayoutInput    textinput.Model
}

func NewModel(projects []scanner.Project, cfg *config.Config, cwd string) Model {
	tools := AllTools(cfg)
	layouts := AllLayouts(cfg)

	m := Model{
		projects: projects,
		cfg:      cfg,
		cwd:      cwd,
		tools:    tools,
		layouts:  layouts,
	}

	// Start on presets if any exist, otherwise mode selection (if >= 2 projects) or project
	if len(cfg.Presets) > 0 {
		m.currentStep = stepPreset
		m.list = newPresetList(cfg.Presets, 60, 20)
	} else if len(projects) >= 2 {
		m.currentStep = stepMode
		m.list = newModeList(60, 20)
	} else {
		m.currentStep = stepProject
		m.list = newProjectList(projects, 60, 20)
	}

	// Prepare text input for preset naming
	ti := textinput.New()
	ti.Placeholder = "my-preset"
	ti.CharLimit = 30
	ti.Width = 30
	m.presetInput = ti

	// Prepare text input for custom layout
	cli := textinput.New()
	cli.Placeholder = "3,4"
	cli.CharLimit = 20
	cli.Width = 20
	m.customLayoutInput = cli

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := appStyle.GetFrameSize()
		listW := msg.Width - h
		overhead := 6 + m.selectionLineCount()
		listH := msg.Height - v - overhead
		if listH < 5 {
			listH = 5
		}
		m.list.SetSize(listW, listH)
		return m, nil

	case tea.KeyMsg:
		// Handle preset naming mode separately
		if m.namingPreset {
			return m.updatePresetNaming(msg)
		}
		// Handle custom layout input mode
		if m.enteringCustomLayout {
			return m.updateCustomLayoutInput(msg)
		}

		// Don't intercept keys when the list is filtering
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit

		case "esc":
			return m.goBack()

		case "enter":
			return m.advance()
		}
	}

	// Delegate to the list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.currentStep == stepDone {
		return ""
	}

	var b strings.Builder

	// Header
	header := headerStyle.Render(titleStyle.Render("Agent T") + "  Workspace Launcher")
	b.WriteString(header)
	b.WriteString("\n")

	// Step indicator
	if m.currentStep != stepPreset {
		num, total := stepNumber(m.currentStep, len(m.cfg.Presets) > 0, m.splitMode)
		b.WriteString(stepStyle.Render(fmt.Sprintf("Step %d/%d: %s", num, total, stepTitle(m.currentStep, m.splitMode))))
		b.WriteString("\n")
	} else {
		b.WriteString(stepStyle.Render(stepTitle(m.currentStep, m.splitMode)))
		b.WriteString("\n")
	}

	// Previous selections summary
	b.WriteString(m.selectionSummary())

	// Preset naming overlay
	if m.namingPreset {
		b.WriteString(m.presetNamingView())
		return appStyle.Render(b.String())
	}

	// Custom layout input overlay
	if m.enteringCustomLayout {
		b.WriteString(m.customLayoutView())
		return appStyle.Render(b.String())
	}

	// Confirm step has a special view
	if m.currentStep == stepConfirm {
		b.WriteString(m.confirmView())
		return appStyle.Render(b.String())
	}

	// List
	b.WriteString(m.list.View())

	return appStyle.Render(b.String())
}

func (m Model) selectionSummary() string {
	var b strings.Builder

	if m.splitMode {
		if m.currentStep > stepProject {
			b.WriteString(selectionLabelStyle.Render("Top:"))
			b.WriteString(selectionValueStyle.Render(m.selectedProject.Name))
			b.WriteString("\n")
		}
		if m.currentStep > stepProjectBottom {
			b.WriteString(selectionLabelStyle.Render("Bottom:"))
			b.WriteString(selectionValueStyle.Render(m.selectedBottomProject.Name))
			b.WriteString("\n")
		}
		if m.currentStep > stepLayout {
			b.WriteString(selectionLabelStyle.Render("Layout:"))
			b.WriteString(selectionValueStyle.Render(fmt.Sprintf("%s (%s)", m.selectedLayout.Name, m.selectedLayout.Desc)))
			b.WriteString("\n")
		}
		if m.currentStep > stepTool {
			b.WriteString(selectionLabelStyle.Render("Top Tool:"))
			b.WriteString(selectionValueStyle.Render(m.selectedTool.Name))
			b.WriteString("\n")
		}
		if m.currentStep > stepToolBottom {
			b.WriteString(selectionLabelStyle.Render("Btm Tool:"))
			b.WriteString(selectionValueStyle.Render(m.selectedToolBottom.Name))
			b.WriteString("\n")
		}
	} else {
		if m.currentStep > stepProject {
			b.WriteString(selectionLabelStyle.Render("Project:"))
			b.WriteString(selectionValueStyle.Render(m.selectedProject.Name))
			b.WriteString("\n")
		}
		if m.currentStep > stepLayout {
			b.WriteString(selectionLabelStyle.Render("Layout:"))
			b.WriteString(selectionValueStyle.Render(fmt.Sprintf("%s (%s)", m.selectedLayout.Name, m.selectedLayout.Desc)))
			b.WriteString("\n")
		}
		if m.currentStep > stepTool {
			b.WriteString(selectionLabelStyle.Render("Tool:"))
			b.WriteString(selectionValueStyle.Render(m.selectedTool.Name))
			b.WriteString("\n")
		}
	}

	if b.Len() > 0 {
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) confirmView() string {
	var b strings.Builder

	// Summary box
	var summary string
	if m.splitMode {
		summary = lipgloss.JoinVertical(lipgloss.Left,
			confirmLabelStyle.Render("Top Project:")+confirmValueStyle.Render(m.selectedProject.Name),
			confirmLabelStyle.Render("Top Dir:")+confirmValueStyle.Render(m.selectedProject.Path),
			confirmLabelStyle.Render("Top Tool:")+confirmValueStyle.Render(m.selectedTool.Name),
			confirmLabelStyle.Render("Btm Project:")+confirmValueStyle.Render(m.selectedBottomProject.Name),
			confirmLabelStyle.Render("Btm Dir:")+confirmValueStyle.Render(m.selectedBottomProject.Path),
			confirmLabelStyle.Render("Btm Tool:")+confirmValueStyle.Render(m.selectedToolBottom.Name),
			confirmLabelStyle.Render("Layout:")+confirmValueStyle.Render(fmt.Sprintf("%s (%s)", m.selectedLayout.Name, m.selectedLayout.Desc)),
		)
	} else {
		summary = lipgloss.JoinVertical(lipgloss.Left,
			confirmLabelStyle.Render("Project:")+confirmValueStyle.Render(m.selectedProject.Name),
			confirmLabelStyle.Render("Layout:")+confirmValueStyle.Render(fmt.Sprintf("%s (%s)", m.selectedLayout.Name, m.selectedLayout.Desc)),
			confirmLabelStyle.Render("Tool:")+confirmValueStyle.Render(m.selectedTool.Name),
			confirmLabelStyle.Render("Directory:")+confirmValueStyle.Render(m.selectedProject.Path),
		)
	}
	b.WriteString(confirmBoxStyle.Render(summary))
	b.WriteString("\n\n")

	// Action list
	b.WriteString(m.list.View())

	return b.String()
}

func (m Model) presetNamingView() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(promptStyle.Render("Preset name: "))
	b.WriteString(m.presetInput.View())
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("Enter to save • Esc to cancel"))
	return b.String()
}

func (m Model) advance() (tea.Model, tea.Cmd) {
	switch m.currentStep {
	case stepPreset:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}
		item := selected.(presetItem)
		if item.isNew {
			// Go to mode selection if >= 2 projects, else straight to project
			if len(m.projects) >= 2 {
				m.currentStep = stepMode
				w, h := m.listSize()
				m.list = newModeList(w, h)
			} else {
				m.currentStep = stepProject
				w, h := m.listSize()
				m.list = newProjectList(m.projects, w, h)
			}
		} else {
			// Apply preset and launch
			m.selectedPreset = &item.preset
			m.applyPreset(item.preset)
			m.currentStep = stepDone
			return m, tea.Quit
		}

	case stepMode:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}
		m.splitMode = selected.(modeItem).split
		m.currentStep = stepProject
		w, h := m.listSize()
		m.list = newProjectList(m.projects, w, h)

	case stepProject:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}
		m.selectedProject = selected.(projectItem).project
		if m.splitMode {
			m.currentStep = stepProjectBottom
			w, h := m.listSize()
			m.list = newProjectList(m.projects, w, h)
		} else {
			m.currentStep = stepLayout
			w, h := m.listSize()
			m.list = newLayoutList(m.layouts, w, h, m.cfg.DefaultLayout)
		}

	case stepProjectBottom:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}
		m.selectedBottomProject = selected.(projectItem).project
		m.currentStep = stepLayout
		w, h := m.listSize()
		m.list = newSplitLayoutList(w, h, m.cfg.DefaultLayout)

	case stepLayout:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}
		lay := selected.(layoutItem).layout
		if lay.Name == "Custom..." {
			m.enteringCustomLayout = true
			cmd := m.customLayoutInput.Focus()
			return m, cmd
		}
		m.selectedLayout = lay
		m.currentStep = stepTool
		w, h := m.listSize()
		m.list = newToolList(m.tools, w, h, m.cfg.DefaultTool)

	case stepTool:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}
		m.selectedTool = selected.(toolItem).tool
		if m.splitMode {
			m.currentStep = stepToolBottom
			w, h := m.listSize()
			m.list = newToolList(m.tools, w, h, m.cfg.DefaultTool)
		} else {
			m.currentStep = stepConfirm
			w, h := m.listSize()
			m.list = newConfirmList(w, h)
		}

	case stepToolBottom:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}
		m.selectedToolBottom = selected.(toolItem).tool
		m.currentStep = stepConfirm
		w, h := m.listSize()
		m.list = newConfirmList(w, h)

	case stepConfirm:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}
		item := selected.(confirmItem)
		if item.name == "Launch" {
			m.currentStep = stepDone
			return m, tea.Quit
		}
		// "Save as preset & Launch"
		m.namingPreset = true
		cmd := m.presetInput.Focus()
		return m, cmd
	}

	return m, nil
}

func (m Model) goBack() (tea.Model, tea.Cmd) {
	switch m.currentStep {
	case stepPreset:
		m.cancelled = true
		return m, tea.Quit

	case stepMode:
		if len(m.cfg.Presets) > 0 {
			m.currentStep = stepPreset
			w, h := m.listSize()
			m.list = newPresetList(m.cfg.Presets, w, h)
		} else {
			m.cancelled = true
			return m, tea.Quit
		}

	case stepProject:
		if len(m.projects) >= 2 {
			m.currentStep = stepMode
			w, h := m.listSize()
			m.list = newModeList(w, h)
		} else if len(m.cfg.Presets) > 0 {
			m.currentStep = stepPreset
			w, h := m.listSize()
			m.list = newPresetList(m.cfg.Presets, w, h)
		} else {
			m.cancelled = true
			return m, tea.Quit
		}

	case stepProjectBottom:
		m.currentStep = stepProject
		w, h := m.listSize()
		m.list = newProjectList(m.projects, w, h)

	case stepLayout:
		if m.splitMode {
			m.currentStep = stepProjectBottom
			w, h := m.listSize()
			m.list = newProjectList(m.projects, w, h)
		} else {
			m.currentStep = stepProject
			w, h := m.listSize()
			m.list = newProjectList(m.projects, w, h)
		}

	case stepTool:
		m.currentStep = stepLayout
		w, h := m.listSize()
		if m.splitMode {
			m.list = newSplitLayoutList(w, h, m.cfg.DefaultLayout)
		} else {
			m.list = newLayoutList(m.layouts, w, h, m.cfg.DefaultLayout)
		}

	case stepToolBottom:
		m.currentStep = stepTool
		w, h := m.listSize()
		m.list = newToolList(m.tools, w, h, m.cfg.DefaultTool)

	case stepConfirm:
		if m.splitMode {
			m.currentStep = stepToolBottom
			w, h := m.listSize()
			m.list = newToolList(m.tools, w, h, m.cfg.DefaultTool)
		} else {
			m.currentStep = stepTool
			w, h := m.listSize()
			m.list = newToolList(m.tools, w, h, m.cfg.DefaultTool)
		}
	}

	return m, nil
}

func (m Model) updatePresetNaming(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(m.presetInput.Value())
		if name == "" {
			return m, nil
		}
		// Save preset
		preset := config.Preset{
			Name:    name,
			Project: m.selectedProject.Name,
			Layout:  m.selectedLayout.ID(),
			Tool:    m.selectedTool.Name,
		}
		if m.splitMode {
			preset.ProjectBottom = m.selectedBottomProject.Name
			preset.ToolBottom = m.selectedToolBottom.Name
		}
		m.cfg.Presets = append(m.cfg.Presets, preset)
		m.configDirty = true
		m.currentStep = stepDone
		return m, tea.Quit

	case "esc":
		m.namingPreset = false
		m.presetInput.Reset()
		return m, nil
	}

	var cmd tea.Cmd
	m.presetInput, cmd = m.presetInput.Update(msg)
	return m, cmd
}

func (m *Model) applyPreset(p config.Preset) {
	// Find the project by name
	for _, proj := range m.projects {
		if proj.Name == p.Project {
			m.selectedProject = proj
			break
		}
	}
	// Normalize old-style layout IDs (e.g. "2x1" -> "2", "3x2" -> "3,3")
	layoutID := convertLegacyLayoutID(p.Layout)
	// Find the layout — search both regular and split layouts
	for _, lay := range m.layouts {
		if lay.ID() == layoutID {
			m.selectedLayout = lay
			break
		}
	}
	if m.selectedLayout.Name == "" {
		for _, lay := range SplitLayouts {
			if lay.ID() == layoutID {
				m.selectedLayout = lay
				break
			}
		}
	}
	// Find the tool
	for _, t := range m.tools {
		if t.Name == p.Tool {
			m.selectedTool = t
			break
		}
	}
	// Detect split preset
	if p.ProjectBottom != "" {
		m.splitMode = true
		for _, proj := range m.projects {
			if proj.Name == p.ProjectBottom {
				m.selectedBottomProject = proj
				break
			}
		}
		for _, t := range m.tools {
			if t.Name == p.ToolBottom {
				m.selectedToolBottom = t
				break
			}
		}
	}
}

// convertLegacyLayoutID converts old "CxR" format to new comma-separated format.
// e.g. "2x1" -> "2", "2x2" -> "2,2", "3x2" -> "3,3", "4x2" -> "4,4"
// New format IDs like "3,4" pass through unchanged.
func convertLegacyLayoutID(id string) string {
	if !strings.Contains(id, "x") {
		return id
	}
	parts := strings.SplitN(id, "x", 2)
	if len(parts) != 2 {
		return id
	}
	cols, err1 := strconv.Atoi(parts[0])
	rows, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return id
	}
	rowCols := make([]string, rows)
	for i := range rowCols {
		rowCols[i] = strconv.Itoa(cols)
	}
	return strings.Join(rowCols, ",")
}

func (m Model) updateCustomLayoutInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		input := strings.TrimSpace(m.customLayoutInput.Value())
		if input == "" {
			return m, nil
		}
		rowCols, err := parseRowCols(input)
		if err != nil {
			return m, nil // ignore invalid input
		}
		layout := Layout{RowCols: rowCols}
		layout.Desc = layout.GenerateDesc()
		layout.Name = fmt.Sprintf("Custom %s", layout.ID())
		m.selectedLayout = layout

		// Save to config
		m.cfg.CustomLayouts = append(m.cfg.CustomLayouts, config.CustomLayout{
			Name:    layout.Name,
			RowCols: rowCols,
		})
		m.configDirty = true
		// Refresh layouts list
		m.layouts = AllLayouts(m.cfg)

		m.enteringCustomLayout = false
		m.customLayoutInput.Reset()
		m.currentStep = stepTool
		w, h := m.listSize()
		m.list = newToolList(m.tools, w, h, m.cfg.DefaultTool)
		return m, nil

	case "esc":
		m.enteringCustomLayout = false
		m.customLayoutInput.Reset()
		return m, nil
	}

	var cmd tea.Cmd
	m.customLayoutInput, cmd = m.customLayoutInput.Update(msg)
	return m, cmd
}

func (m Model) customLayoutView() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(promptStyle.Render("Columns per row: "))
	b.WriteString(m.customLayoutInput.View())
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("e.g. 3,4 = 3 top, 4 bottom"))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("Enter to confirm • Esc to cancel"))
	return b.String()
}

// parseRowCols parses a comma-separated string like "3,4" into []int{3, 4}.
func parseRowCols(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	total := 0
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 {
			return nil, fmt.Errorf("invalid: %s", p)
		}
		result = append(result, n)
		total += n
	}
	if len(result) == 0 || total > 20 {
		return nil, fmt.Errorf("invalid layout")
	}
	return result, nil
}

func (m Model) selectionLineCount() int {
	count := 0
	if m.splitMode {
		if m.currentStep > stepProject {
			count++ // "Top: xxx"
		}
		if m.currentStep > stepProjectBottom {
			count++ // "Bottom: xxx"
		}
		if m.currentStep > stepLayout {
			count++ // "Layout: xxx"
		}
		if m.currentStep > stepTool {
			count++ // "Top Tool: xxx"
		}
		if m.currentStep > stepToolBottom {
			count++ // "Btm Tool: xxx"
		}
	} else {
		if m.currentStep > stepProject {
			count++ // "Project: xxx"
		}
		if m.currentStep > stepLayout {
			count++ // "Layout: xxx"
		}
		if m.currentStep > stepTool {
			count++ // "Tool: xxx"
		}
	}
	if count > 0 {
		count++ // blank line after selections
	}
	return count
}

func (m Model) listSize() (int, int) {
	h, v := appStyle.GetFrameSize()
	w := m.width - h
	overhead := 6 + m.selectionLineCount()
	lh := m.height - v - overhead
	if w < 30 {
		w = 60
	}
	if lh < 5 {
		lh = 20
	}
	return w, lh
}

// --- Accessors for main.go ---

func (m Model) Cancelled() bool                    { return m.cancelled }
func (m Model) ConfigChanged() bool                 { return m.configDirty }
func (m Model) Config() *config.Config              { return m.cfg }
func (m Model) SelectedProject() scanner.Project    { return m.selectedProject }
func (m Model) SelectedLayout() Layout              { return m.selectedLayout }
func (m Model) SelectedTool() Tool                  { return m.selectedTool }
func (m Model) IsSplitMode() bool                   { return m.splitMode }
func (m Model) SelectedBottomProject() scanner.Project { return m.selectedBottomProject }
func (m Model) SelectedToolBottom() Tool            { return m.selectedToolBottom }
