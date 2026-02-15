package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultLayout  string            `yaml:"default_layout,omitempty"`
	DefaultTool    string            `yaml:"default_tool,omitempty"`
	CustomCommands map[string]string `yaml:"custom_commands,omitempty"`
	Presets        []Preset          `yaml:"presets,omitempty"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "agent-t", "config.yaml")
}

func Load() (*Config, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				CustomCommands: make(map[string]string),
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.CustomCommands == nil {
		cfg.CustomCommands = make(map[string]string)
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
