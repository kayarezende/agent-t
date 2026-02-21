package launcher

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
)

type Options struct {
	ProjectDirs []string // one project dir per row
	RowCols     []int    // columns per row, e.g. [3,4] = 3 top, 4 bottom
	Commands    []string // one tool command per row (empty string = no tool)
}

type screenBounds struct {
	X1, Y1, X2, Y2 int
}

type scriptData struct {
	X1, Y1, X2, Y2 int
	RowCols         string // comma-separated, e.g. "3,4"
	NumRows         int
	TermCmds        string // AppleScript list literal, e.g. "\"cd ... && claude\", \"cd ... && codex\""
}

func Launch(opts Options) error {
	bounds, err := detectScreen()
	if err != nil {
		// Fallback to 1920x1080
		bounds = screenBounds{X1: 0, Y1: 0, X2: 1920, Y2: 1080}
	}

	// Build per-row terminal commands
	numRows := len(opts.RowCols)
	termCmds := make([]string, numRows)
	for r := 0; r < numRows; r++ {
		dir := opts.ProjectDirs[0]
		if r < len(opts.ProjectDirs) {
			dir = opts.ProjectDirs[r]
		}
		cmd := ""
		if r < len(opts.Commands) {
			cmd = opts.Commands[r]
		}
		termCmds[r] = fmt.Sprintf("cd %s && clear", shellQuote(dir))
		if cmd != "" {
			termCmds[r] += " && " + cmd
		}
	}

	script, err := buildTilingScript(bounds, opts.RowCols, termCmds)
	if err != nil {
		return fmt.Errorf("building AppleScript: %w", err)
	}

	return execAppleScript(script)
}

func detectScreen() (screenBounds, error) {
	cmd := exec.Command("osascript", "-l", "JavaScript", "-e", jxaScreenDetect)
	out, err := cmd.Output()
	if err != nil {
		return screenBounds{}, fmt.Errorf("screen detection failed: %w", err)
	}

	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 4 {
		return screenBounds{}, fmt.Errorf("unexpected screen detection output: %q", string(out))
	}

	vals := make([]int, 4)
	for i, p := range parts {
		v, err := strconv.Atoi(p)
		if err != nil {
			return screenBounds{}, fmt.Errorf("parsing screen bound %q: %w", p, err)
		}
		vals[i] = v
	}

	return screenBounds{X1: vals[0], Y1: vals[1], X2: vals[2], Y2: vals[3]}, nil
}

func buildTilingScript(bounds screenBounds, rowCols []int, termCmds []string) (string, error) {
	tmpl, err := template.New("tiling").Parse(tilingScriptTemplate)
	if err != nil {
		return "", err
	}

	// Build comma-separated rowCols string for AppleScript
	parts := make([]string, len(rowCols))
	for i, c := range rowCols {
		parts[i] = strconv.Itoa(c)
	}

	// Build AppleScript list literal for per-row commands
	// Each command is double-quote-escaped for AppleScript string embedding
	cmdParts := make([]string, len(termCmds))
	for i, cmd := range termCmds {
		escaped := strings.ReplaceAll(cmd, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		cmdParts[i] = fmt.Sprintf("\"%s\"", escaped)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, scriptData{
		X1:       bounds.X1,
		Y1:       bounds.Y1,
		X2:       bounds.X2,
		Y2:       bounds.Y2,
		RowCols:  strings.Join(parts, ", "),
		NumRows:  len(rowCols),
		TermCmds: strings.Join(cmdParts, ", "),
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func execAppleScript(script string) error {
	cmd := exec.Command("osascript", "-e", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("osascript error: %w\noutput: %s", err, string(out))
	}
	return nil
}

// shellQuote wraps a string in single quotes for safe shell embedding.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
