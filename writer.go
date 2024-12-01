package main

import (
	"fmt"
	"os"
	"time"

	"github.com/hans-m-song/watchdog/pkg/colors"
	"github.com/hans-m-song/watchdog/pkg/daemon"
)

func writeStdout(d daemon.Daemon, msg string) {
	fmt.Fprintf(
		os.Stdout,
		"%s %s > %s\n",
		colors.Dim.Surround(time.Now().Format(time.Kitchen)),
		colors.Bold.Surround(colors.RandD(d.Name()).Surround(d.ID())),
		msg,
	)
}

func writeStderr(d daemon.Daemon, msg string) {
	fmt.Fprintf(
		os.Stderr,
		"%s %s > %s\n",
		colors.Dim.Surround(time.Now().Format(time.Kitchen)),
		colors.Bold.Surround(colors.Red.Surround(d.ID())),
		msg,
	)
}
