package daemon_test

import (
	"testing"
	"time"

	"github.com/hans-m-song/watchdog/pkg/daemon"
	"github.com/stretchr/testify/assert"
)

func TestDaemonLifecycle(t *testing.T) {
	d, err := daemon.New(daemon.Config{Name: "test", Command: "sleep 1s"})
	assert.NoError(t, err)
	assert.Equal(t, "test", d.Name())

	t.Cleanup(func() {
		err := d.Stop()
		assert.NoError(t, err)
	})

	err = d.Start()
	assert.NoError(t, err)

	err = d.Reload()
	assert.NoError(t, err)
}

func TestDaemonCallback(t *testing.T) {
	d, err := daemon.New(daemon.Config{
		Name:    "test",
		Command: "echo 'stdout' && echo 'stderr' >&2",
	})
	assert.NoError(t, err)

	t.Cleanup(func() {
		if err := d.Stop(); err != nil {
			assert.NoError(t, err)
		}
	})

	stdout := ""
	d.OnStdout(func(d daemon.Daemon, msg string) { stdout = msg })

	stderr := ""
	d.OnStderr(func(d daemon.Daemon, msg string) { stderr = msg })

	err = d.Start()
	assert.NoError(t, err)

	assert.Eventually(
		t,
		func() bool { return stdout == "stdout" },
		time.Second,
		100*time.Millisecond,
	)

	assert.Eventually(
		t,
		func() bool { return stderr == "stderr" },
		time.Second,
		100*time.Millisecond,
	)
}
