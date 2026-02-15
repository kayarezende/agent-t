# Agent T

A terminal workspace launcher for macOS. Pick a project, choose a layout, optionally launch an AI coding tool, and Agent T tiles Terminal.app windows across your screen.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for a polished interactive TUI.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
![macOS](https://img.shields.io/badge/macOS-only-000000?logo=apple&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-blue)

<img width="457" height="229" alt="image" src="https://github.com/user-attachments/assets/59269976-3348-447b-ba4b-2b5f255da3fe" />

<img width="459" height="480" alt="image" src="https://github.com/user-attachments/assets/e39e9888-62c4-418b-969d-34eccfc16325" />

<img width="627" height="525" alt="image" src="https://github.com/user-attachments/assets/184d0e9d-242a-4d38-bd6a-bf6d98cd4c92" />

<img width="2553" height="1349" alt="image" src="https://github.com/user-attachments/assets/2854088d-a2f8-4bc1-b5fe-170ecd23d365" />



## Features
<img width="433" height="232" alt="image" src="https://github.com/user-attachments/assets/ae385713-2c98-45f6-a465-10a113db10c9" />

- **Fuzzy project search** — type to instantly filter your project list
- **Terminal tiling** — automatically opens and tiles 2, 4, 6, or 8 Terminal.app windows
- **AI tool integration** — launch Claude Code, Codex, OpenCode, or any custom command in every terminal
- **Saved presets** — save your favorite project + layout + tool combos for one-key launch
- **Custom commands** — define your own tools (Cursor, Vim, Zed, etc.) in the config file
- **Multi-monitor support** — detects which screen your terminal is on and tiles there
- **Back navigation** — press Esc to go back a step, Ctrl+C to quit

## Install

### From source (requires Go 1.22+)

```bash
git clone https://github.com/kayarezende/agent-t.git
cd agent-t
go build -o agent-t .
```

Then add it to your PATH so you can run `agent-t` from any terminal:

```bash
# Create a local bin directory and add to PATH
mkdir -p ~/bin
cp agent-t ~/bin/
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

Or if you prefer `/usr/local/bin`:

```bash
sudo cp agent-t /usr/local/bin/
```

### From releases

Download the latest binary from [Releases](https://github.com/kayarezende/agent-t/releases), then:

```bash
mkdir -p ~/bin
cp agent-t ~/bin/
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

## Usage

Navigate to a directory containing your project folders and run:

```bash
cd ~/GitHub
agent-t
```

Agent T scans the current directory for subdirectories and walks you through:

```
Step 1/4 → Select a project (fuzzy searchable)
Step 2/4 → Pick a terminal layout (2, 4, 6, or 8 terminals)
Step 3/4 → Choose an AI tool (or none)
Step 4/4 → Confirm and launch
```

Terminals are tiled across your screen automatically.

### Keyboard Controls

| Key | Action |
|-----|--------|
| `↑` / `↓` or `j` / `k` | Navigate |
| `Enter` | Select |
| `/` | Filter (on project list) |
| `Esc` | Go back one step |
| `Ctrl+C` | Quit |

## Configuration

Agent T uses a config file at `~/.config/agent-t/config.yaml`. It's created automatically when you save your first preset.

```yaml
# Default selections (pre-selected but changeable)
default_layout: "2x2"
default_tool: "Claude Code"

# Add your own tools alongside the built-in ones
custom_commands:
  Cursor: "cursor ."
  Vim: "nvim"
  Zed: "zed ."

# Saved presets for quick launch
presets:
  - name: "api-claude"
    project: "api-service"
    layout: "2x2"
    tool: "Claude Code"
  - name: "frontend-dev"
    project: "web-app"
    layout: "3x2"
    tool: "None"
```

### Presets

When presets exist, Agent T shows them first on launch. Select a preset to instantly launch that workspace, or choose "New workspace..." to go through the normal flow.

To save a preset, complete the wizard and choose "Save as preset & Launch" on the confirm screen.

### Custom Commands

Add any command to `custom_commands` in the config. The key is the display name, the value is the command to run:

```yaml
custom_commands:
  Cursor: "cursor ."
  "My Script": "bash ~/scripts/dev-setup.sh"
```

Custom commands appear alongside the built-in tools in the tool selection step.

## Built-in Tools

| Tool | Command |
|------|---------|
| None | Just opens terminals |
| Claude Code | `claude` |
| Claude Code (Chrome) | `claude --chat-mode browser` |
| Codex | `codex` |
| OpenCode | `opencode` |

## Layouts

| Layout | Grid | Terminals |
|--------|------|-----------|
| 2 terminals | 2x1 | `[ ][ ]` |
| 4 terminals | 2x2 | `[ ][ ]` / `[ ][ ]` |
| 6 terminals | 3x2 | `[ ][ ][ ]` / `[ ][ ][ ]` |
| 8 terminals | 4x2 | `[ ][ ][ ][ ]` / `[ ][ ][ ][ ]` |

## Requirements

- **macOS** (uses AppleScript to control Terminal.app)
- **Go 1.22+** (to build from source)
- **Terminal.app** (the default macOS terminal)

## How It Works

1. Scans the current directory for project subdirectories
2. Presents an interactive TUI wizard using Bubble Tea
3. Detects which screen your terminal is on via JXA (JavaScript for Automation)
4. Generates and executes AppleScript to open Terminal.app windows
5. Tiles the windows in a grid layout across your screen

## Project Structure

```
├── main.go                  # Entry point
├── internal/
│   ├── tui/                 # Bubble Tea UI
│   │   ├── model.go         # Main model (Init/Update/View)
│   │   ├── steps.go         # Wizard steps and data types
│   │   └── styles.go        # Lipgloss styling
│   ├── config/              # YAML config management
│   │   ├── config.go        # Load/Save config
│   │   └── preset.go        # Preset type
│   ├── scanner/             # Directory scanning
│   │   └── scanner.go       # Scan for projects
│   └── launcher/            # Terminal tiling
│       ├── launcher.go      # Screen detection + launch
│       └── scripts.go       # AppleScript templates
├── go.mod
└── go.sum
```

## License

MIT
