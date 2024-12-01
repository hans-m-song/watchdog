package daemon

import (
	"bufio"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/hans-m-song/watchdog/pkg/flow"
	"github.com/rs/zerolog/log"
)

type Daemon interface {
	Name() string
	ID() string
	Match(string) bool

	OnStdout(func(Daemon, string))
	OnStderr(func(Daemon, string))

	Start() error
	Stop() error
	Reload() error
}

type daemonImpl struct {
	name          string
	command       string
	globs         []glob.Glob
	restartOnExit bool
	restartDelay  time.Duration

	childMu sync.Mutex
	child   *exec.Cmd

	debounce        func(func())
	callbacksMu     sync.Mutex
	stdoutCallbacks []func(Daemon, string)
	stderrCallbacks []func(Daemon, string)
}

func New(config Config) (Daemon, error) {
	d := daemonImpl{
		name:          config.Name,
		command:       config.Command,
		restartOnExit: config.RestartOnExit,
		restartDelay:  time.Duration(config.RestartDelay),
		debounce:      flow.NewDebouncer(1 * time.Second),
	}

	if config.RestartDelay <= 0 {
		d.restartDelay = time.Second
	}

	for _, path := range config.Paths {
		glob, err := glob.Compile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to compile path '%s' to glob: %w", path, err)
		}

		d.globs = append(d.globs, glob)
	}

	return &d, nil
}

func (d *daemonImpl) Name() string {
	return d.name
}

func (d *daemonImpl) ID() string {
	if d.child == nil || d.child.Process == nil {
		return fmt.Sprintf("%s:stopped", d.name)
	}

	return fmt.Sprintf("%s:%d", d.name, d.child.Process.Pid)
}

func (d *daemonImpl) Match(changed string) bool {
	for _, glob := range d.globs {
		if glob.Match(changed) {
			return true
		}
	}

	return false
}

func (d *daemonImpl) Start() error {
	d.childMu.Lock()
	defer d.childMu.Unlock()

	d.child = exec.Command("bash", "-c", d.command)

	var err error

	stdout, err := d.child.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to pipe stdout: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			for _, callback := range d.stdoutCallbacks {
				go callback(d, scanner.Text())
			}
		}
	}()

	stderr, err := d.child.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to pipe stderr: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			for _, callback := range d.stderrCallbacks {
				go callback(d, scanner.Text())
			}
		}
	}()

	log.Debug().Str("daemon", d.name).Str("command", d.command).Msg("starting")
	if err := d.child.Start(); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	go func() {
		log := log.With().Str("daemon", d.ID()).Logger()

		if d.child == nil || d.child.ProcessState != nil {
			return
		}

		if err := d.child.Wait(); err != nil {
			log.Warn().Err(err).Int("exit_code", d.child.ProcessState.ExitCode()).Msg("exited")
		} else {
			log.Debug().Msg("exited cleanly")
		}

		d.childMu.Lock()
		d.child = nil
		d.childMu.Unlock()

		if d.restartOnExit {
			go d.Start()
		}
	}()

	return nil
}

func (d *daemonImpl) Stop() error {
	d.childMu.Lock()
	defer d.childMu.Unlock()

	if d.child == nil || d.child.ProcessState != nil {
		return nil
	}

	if err := d.child.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill %d: %w", d.child.Process.Pid, err)
	}

	d.child = nil

	return nil
}

func (d *daemonImpl) Reload() error {
	if d.debounce == nil {
		d.debounce = flow.NewDebouncer(1 * time.Second)
	}

	d.debounce(func() {
		log := log.With().Str("daemon", d.ID()).Logger()

		log.Debug().Msg("reloading")

		if err := d.Stop(); err != nil {
			log.Error().Err(err).Send()
			return
		}

		if err := d.Start(); err != nil {
			log.Error().Err(err).Send()
			return
		}
	})

	return nil
}

func (d *daemonImpl) OnStdout(callback func(Daemon, string)) {
	d.callbacksMu.Lock()
	defer d.callbacksMu.Unlock()

	d.stdoutCallbacks = append(d.stdoutCallbacks, callback)
}

func (d *daemonImpl) OnStderr(callback func(Daemon, string)) {
	d.callbacksMu.Lock()
	defer d.callbacksMu.Unlock()

	d.stderrCallbacks = append(d.stderrCallbacks, callback)
}
