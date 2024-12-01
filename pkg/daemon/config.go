package daemon

import (
	"github.com/hans-m-song/watchdog/pkg/serde"
)

type Config struct {
	Name          string         `yaml:"name"`
	Command       string         `yaml:"command"`
	Paths         []string       `yaml:"paths"`
	RestartOnExit bool           `yaml:"restart_on_exit"`
	RestartDelay  serde.Duration `yaml:"restart_delay"`
}
