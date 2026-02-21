package tui

import (
	"testing"
)

func TestStepTitle_SingleMode(t *testing.T) {
	tests := []struct {
		step step
		want string
	}{
		{stepProject, "Select Project"},
		{stepTool, "Select AI Tool"},
		{stepMode, "Workspace Mode"},
		{stepConfirm, "Confirm & Launch"},
	}
	for _, tt := range tests {
		got := stepTitle(tt.step, false)
		if got != tt.want {
			t.Errorf("stepTitle(%d, false) = %q, want %q", tt.step, got, tt.want)
		}
	}
}

func TestStepTitle_SplitMode(t *testing.T) {
	tests := []struct {
		step step
		want string
	}{
		{stepProject, "Select Top Project"},
		{stepProjectBottom, "Select Bottom Project"},
		{stepTool, "Select Top Tool"},
		{stepToolBottom, "Select Bottom Tool"},
	}
	for _, tt := range tests {
		got := stepTitle(tt.step, true)
		if got != tt.want {
			t.Errorf("stepTitle(%d, true) = %q, want %q", tt.step, got, tt.want)
		}
	}
}

func TestStepNumber_SingleMode(t *testing.T) {
	num, total := stepNumber(stepProject, false, false)
	if total != 4 {
		t.Errorf("single mode total = %d, want 4", total)
	}
	if num != 1 {
		t.Errorf("stepProject number = %d, want 1", num)
	}

	num, total = stepNumber(stepConfirm, false, false)
	if num != 4 || total != 4 {
		t.Errorf("stepConfirm = %d/%d, want 4/4", num, total)
	}
}

func TestStepNumber_SplitMode(t *testing.T) {
	num, total := stepNumber(stepProject, false, true)
	if total != 7 {
		t.Errorf("split mode total = %d, want 7", total)
	}
	if num != 2 {
		t.Errorf("stepProject in split = %d, want 2", num)
	}

	num, _ = stepNumber(stepProjectBottom, false, true)
	if num != 3 {
		t.Errorf("stepProjectBottom = %d, want 3", num)
	}

	num, _ = stepNumber(stepToolBottom, false, true)
	if num != 6 {
		t.Errorf("stepToolBottom = %d, want 6", num)
	}

	num, total = stepNumber(stepConfirm, false, true)
	if num != 7 || total != 7 {
		t.Errorf("stepConfirm in split = %d/%d, want 7/7", num, total)
	}
}

func TestSplitLayouts(t *testing.T) {
	if len(SplitLayouts) == 0 {
		t.Fatal("SplitLayouts should not be empty")
	}
	for _, lay := range SplitLayouts {
		if len(lay.RowCols) != 2 {
			t.Errorf("SplitLayout %q has %d rows, want exactly 2", lay.Name, len(lay.RowCols))
		}
	}
}

func TestModeItem(t *testing.T) {
	single := modeItem{name: "Single Project", desc: "All terminals in one project", split: false}
	split := modeItem{name: "Split Workspace", desc: "Top row = project A, bottom row = project B", split: true}

	if single.Title() != "Single Project" {
		t.Errorf("single Title() = %q", single.Title())
	}
	if split.Title() != "Split Workspace" {
		t.Errorf("split Title() = %q", split.Title())
	}
	if single.split {
		t.Error("single mode should not be split")
	}
	if !split.split {
		t.Error("split mode should be split")
	}
}
