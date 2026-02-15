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
	ProjectDir string
	Cols       int
	Rows       int
	Command    string
}

type screenBounds struct {
	X1, Y1, X2, Y2 int
}

type scriptData struct {
	X1, Y1, X2, Y2 int
	Cols, Rows      int
	TermCmd         string
}

func Launch(opts Options) error {
	bounds, err := detectScreen()
	if err != nil {
		// Fallback to 1920x1080
		bounds = screenBounds{X1: 0, Y1: 0, X2: 1920, Y2: 1080}
	}

	termCmd := fmt.Sprintf("cd %s && clear", shellQuote(opts.ProjectDir))
	if opts.Command != "" {
		termCmd += " && " + opts.Command
	}

	script, err := buildTilingScript(bounds, opts.Cols, opts.Rows, termCmd)
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

func buildTilingScript(bounds screenBounds, cols, rows int, termCmd string) (string, error) {
	tmpl, err := template.New("tiling").Parse(tilingScriptTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, scriptData{
		X1:      bounds.X1,
		Y1:      bounds.Y1,
		X2:      bounds.X2,
		Y2:      bounds.Y2,
		Cols:    cols,
		Rows:    rows,
		TermCmd: termCmd,
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
