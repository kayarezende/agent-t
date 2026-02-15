package tui

import (
	"fmt"
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

	list list.Model

	selectedPreset  *config.Preset
	selectedProject scanner.Project
	selectedLayout  Layout
	selectedTool    Tool

	// Preset naming
	namingPreset bool
	presetInput  textinput.Model
}

func NewModel(projects []scanner.Project, cfg *config.Config, cwd string) Model {
	tools := AllTools(cfg)

	m := Model{
		projects: projects,
		cfg:      cfg,
		cwd:      cwd,
		tools:    tools,
	}

	// Start on presets if any exist, otherwise jump to project selection
	if len(cfg.Presets) > 0 {
		m.currentStep = stepPreset
		m.list = newPresetList(cfg.Presets, 60, 20)
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
		num, total := stepNumber(m.currentStep, len(m.cfg.Presets) > 0)
		b.WriteString(stepStyle.Render(fmt.Sprintf("Step %d/%d: %s", num, total, stepTitle(m.currentStep))))
		b.WriteString("\n")
	} else {
		b.WriteString(stepStyle.Render(stepTitle(m.currentStep)))
		b.WriteString("\n")
	}

	// Previous selections summary
	b.WriteString(m.selectionSummary())

	// Preset naming overlay
	if m.namingPreset {
		b.WriteString(m.presetNamingView())
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
	if m.currentStep > stepProject {
		b.WriteString(selectionLabelStyle.Render("Project:"))
		b.WriteString(selectionValueStyle.Render(m.selectedProject.Name))
		b.WriteString("\n")
	}
	if m.currentStep > stepLayout {
		b.WriteString(selectionLabelStyle.Render("Layout:"))
		b.WriteString(selectionValueStyle.Render(fmt.Sprintf("%s (%dx%d)", m.selectedLayout.Name, m.selectedLayout.Cols, m.selectedLayout.Rows)))
		b.WriteString("\n")
	}
	if m.currentStep > stepTool {
		b.WriteString(selectionLabelStyle.Render("Tool:"))
		b.WriteString(selectionValueStyle.Render(m.selectedTool.Name))
		b.WriteString("\n")
	}
	if b.Len() > 0 {
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) confirmView() string {
	var b strings.Builder

	// Summary box
	summary := lipgloss.JoinVertical(lipgloss.Left,
		confirmLabelStyle.Render("Project:")+confirmValueStyle.Render(m.selectedProject.Name),
		confirmLabelStyle.Render("Layout:")+confirmValueStyle.Render(fmt.Sprintf("%s (%dx%d)", m.selectedLayout.Name, m.selectedLayout.Cols, m.selectedLayout.Rows)),
		confirmLabelStyle.Render("Tool:")+confirmValueStyle.Render(m.selectedTool.Name),
		confirmLabelStyle.Render("Directory:")+confirmValueStyle.Render(m.selectedProject.Path),
	)
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
			m.currentStep = stepProject
			w, h := m.listSize()
			m.list = newProjectList(m.projects, w, h)
		} else {
			// Apply preset and launch
			m.selectedPreset = &item.preset
			m.applyPreset(item.preset)
			m.currentStep = stepDone
			return m, tea.Quit
		}

	case stepProject:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}
		m.selectedProject = selected.(projectItem).project
		m.currentStep = stepLayout
		w, h := m.listSize()
		m.list = newLayoutList(w, h, m.cfg.DefaultLayout)

	case stepLayout:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}
		m.selectedLayout = selected.(layoutItem).layout
		m.currentStep = stepTool
		w, h := m.listSize()
		m.list = newToolList(m.tools, w, h, m.cfg.DefaultTool)

	case stepTool:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}
		m.selectedTool = selected.(toolItem).tool
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

	case stepProject:
		if len(m.cfg.Presets) > 0 {
			m.currentStep = stepPreset
			w, h := m.listSize()
			m.list = newPresetList(m.cfg.Presets, w, h)
		} else {
			m.cancelled = true
			return m, tea.Quit
		}

	case stepLayout:
		m.currentStep = stepProject
		w, h := m.listSize()
		m.list = newProjectList(m.projects, w, h)

	case stepTool:
		m.currentStep = stepLayout
		w, h := m.listSize()
		m.list = newLayoutList(w, h, m.cfg.DefaultLayout)

	case stepConfirm:
		m.currentStep = stepTool
		w, h := m.listSize()
		m.list = newToolList(m.tools, w, h, m.cfg.DefaultTool)
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
	// Find the layout
	for _, lay := range Layouts {
		if lay.ID() == p.Layout {
			m.selectedLayout = lay
			break
		}
	}
	// Find the tool
	for _, t := range m.tools {
		if t.Name == p.Tool {
			m.selectedTool = t
			break
		}
	}
}

func (m Model) selectionLineCount() int {
	count := 0
	if m.currentStep > stepProject {
		count++ // "Project: xxx"
	}
	if m.currentStep > stepLayout {
		count++ // "Layout: xxx"
	}
	if m.currentStep > stepTool {
		count++ // "Tool: xxx"
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

func (m Model) Cancelled() bool          { return m.cancelled }
func (m Model) ConfigChanged() bool       { return m.configDirty }
func (m Model) Config() *config.Config    { return m.cfg }
func (m Model) SelectedProject() scanner.Project { return m.selectedProject }
func (m Model) SelectedLayout() Layout    { return m.selectedLayout }
func (m Model) SelectedTool() Tool        { return m.selectedTool }
