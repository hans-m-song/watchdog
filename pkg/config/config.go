package config

import (
	"fmt"
	"os"

	"github.com/hans-m-song/watchdog/pkg/serde"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Tasks []Task `yaml:"tasks"`
}

func (c Config) Valid() error {
	for i, task := range c.Tasks {
		if err := task.Valid(); err != nil {
			return fmt.Errorf("task %d: %w", i, err)
		}
	}

	return nil
}

type Task struct {
	Name          string         `yaml:"name"`
	Command       string         `yaml:"command"`
	Paths         []string       `yaml:"paths"`
	RestartOnExit bool           `yaml:"restart_on_exit"`
	RestartDelay  serde.Duration `yaml:"restart_delay"`
}

func (t Task) Valid() error {
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}

	if t.Command == "" {
		return fmt.Errorf("command is required")
	}

	return nil
}

func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(raw, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.Valid(); err != nil {
		return nil, err
	}

	return &config, nil
}
