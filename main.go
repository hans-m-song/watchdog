package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/axatol/gonfig"
	"github.com/hans-m-song/watchdog/pkg/config"
	"github.com/hans-m-song/watchdog/pkg/daemon"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Args struct {
	Config   string `flag:"config" env:"CONFIG_FILENAME" default:"watchdog.yaml"`
	LogLevel string `flag:"log-level" env:"LOG_LEVEL" default:"info"`
}

func main() {
	var args Args
	if err := gonfig.Load(&args); err != nil {
		panic(fmt.Errorf("failed to load initialise: %w", err))
	}

	logLevel, err := zerolog.ParseLevel(args.LogLevel)
	if err != nil {
		panic(fmt.Errorf("invalid log level: %w", err))
	}

	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		Level(zerolog.Level(logLevel)).
		With().
		Caller().
		Timestamp().
		Logger()

	cfg, err := config.Load(args.Config)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	log.Info().Any("config", cfg).Send()

	daemons := make([]daemon.Daemon, 0, len(cfg.Tasks))
	defer func() {
		for _, d := range daemons {
			if err := d.Stop(); err != nil {
				log.Error().Err(err).Send()
			}
		}
	}()

	for _, task := range cfg.Tasks {
		d, err := daemon.New(daemon.Config{
			Name:    task.Name,
			Command: task.Command,
			Paths:   task.Paths,

			RestartOnExit: task.RestartOnExit,
			RestartDelay:  task.RestartDelay,
		})

		if err != nil {
			log.Fatal().Err(err).Send()
		}

		d.OnStdout(writeStdout)
		d.OnStderr(writeStderr)

		if err := d.Start(); err != nil {
			log.Fatal().Err(err).Send()
		}

		daemons = append(daemons, d)
	}

	fsd := daemon.FSWatcherDaemon{}
	defer func() {
		if err := fsd.Stop(); err != nil {
			log.Error().Err(err).Send()
		}
	}()

	fsd.OnChange(func(path string) {
		log.Trace().Str("path", path).Msg("change detected")

		for _, d := range daemons {
			log.Trace().Str("daemon", d.ID()).Str("path", path).Bool("match", d.Match(path)).Msg("checking")
			if d.Match(path) {
				log.Debug().Str("daemon", d.ID()).Msg("reloading")
				if err := d.Reload(); err != nil {
					log.Error().Err(err).Send()
				}
			}
		}
	})

	if err := fsd.Start(); err != nil {
		log.Fatal().Err(err).Send()
	}

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	<-ctx.Done()
}
