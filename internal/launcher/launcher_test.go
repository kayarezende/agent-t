package launcher

import (
	"strings"
	"testing"
)

func TestBuildTilingScript_SingleProject(t *testing.T) {
	bounds := screenBounds{X1: 0, Y1: 0, X2: 1920, Y2: 1080}
	rowCols := []int{3, 3}
	termCmds := []string{
		"cd '/projects/api' && clear && claude",
		"cd '/projects/api' && clear && claude",
	}

	script, err := buildTilingScript(bounds, rowCols, termCmds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain the termCmdsList with both commands
	if !strings.Contains(script, "set termCmdsList to") {
		t.Error("script missing termCmdsList declaration")
	}
	if !strings.Contains(script, "set thisCmd to item r of termCmdsList") {
		t.Error("script missing per-row command selection")
	}
	if !strings.Contains(script, "do script thisCmd") {
		t.Error("script missing 'do script thisCmd'")
	}
	// Should NOT contain the old hardcoded pattern
	if strings.Contains(script, "do script \"cd") {
		t.Error("script should not contain hardcoded do script command")
	}
}

func TestBuildTilingScript_SplitProjects(t *testing.T) {
	bounds := screenBounds{X1: 0, Y1: 0, X2: 1920, Y2: 1080}
	rowCols := []int{3, 3}
	termCmds := []string{
		"cd '/projects/api' && clear && claude",
		"cd '/projects/frontend' && clear && codex",
	}

	script, err := buildTilingScript(bounds, rowCols, termCmds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both commands should appear in the script
	if !strings.Contains(script, "/projects/api") {
		t.Error("script missing api project path")
	}
	if !strings.Contains(script, "/projects/frontend") {
		t.Error("script missing frontend project path")
	}
	if !strings.Contains(script, "claude") {
		t.Error("script missing claude command")
	}
	if !strings.Contains(script, "codex") {
		t.Error("script missing codex command")
	}
}

func TestBuildTilingScript_EscapesQuotes(t *testing.T) {
	bounds := screenBounds{X1: 0, Y1: 0, X2: 1920, Y2: 1080}
	rowCols := []int{2}
	termCmds := []string{
		`cd '/projects/my "project"' && clear`,
	}

	script, err := buildTilingScript(bounds, rowCols, termCmds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Double quotes in the command should be escaped for AppleScript
	if !strings.Contains(script, `\"project\"`) {
		t.Error("script should escape double quotes in commands")
	}
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "'simple'"},
		{"path with spaces", "'path with spaces'"},
		{"it's", "'it'\\''s'"},
	}
	for _, tt := range tests {
		got := shellQuote(tt.input)
		if got != tt.want {
			t.Errorf("shellQuote(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
