package daemon

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

type FSWatcherDaemon struct {
	watcher     *fsnotify.Watcher
	callbacksMu sync.Mutex
	callbacks   []func(string)
}

func (d *FSWatcherDaemon) OnChange(callback func(string)) {
	d.callbacksMu.Lock()
	defer d.callbacksMu.Unlock()

	d.callbacks = append(d.callbacks, callback)
}

func (d *FSWatcherDaemon) Start() error {
	if d.watcher != nil {
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	d.watcher = watcher

	go func() {
		for {
			select {
			case event, ok := <-d.watcher.Events:
				if !ok {
					return
				}

				log.Trace().Str("path", event.Name).Str("op", event.Op.String()).Send()
				d.callbacksMu.Lock()
				for _, callback := range d.callbacks {
					go callback(event.Name)
				}
				d.callbacksMu.Unlock()

			case err, ok := <-d.watcher.Errors:
				if !ok {
					return
				}

				log.Error().Err(err).Send()
			}
		}
	}()

	if err := watcher.Add("."); err != nil {
		return fmt.Errorf("failed to start watcher")
	}

	log.Debug().Msg("watcher started")
	return nil
}

func (d *FSWatcherDaemon) Stop() error {
	return d.watcher.Close()
}

// Reload is a no-op
func (d *FSWatcherDaemon) Reload() error {
	return nil
}
