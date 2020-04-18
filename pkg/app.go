package pkg

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/smitajit/logtrics/config"
	"github.com/smitajit/logtrics/pkg/reader"
)

type Application struct {
	reader  reader.LogReader
	scripts []*Script
	ctx     context.Context
	cncel   context.CancelFunc
	logger  zerolog.Logger
}

func NewApp(config *config.Configuration, reader reader.LogReader, scripts ...string) (*Application, error) {
	app := &Application{
		reader:  reader,
		scripts: make([]*Script, 0),
		logger:  config.Logger("application"),
	}
	app.ctx, app.cncel = context.WithCancel(context.Background())
	for _, s := range scripts {
		script, err := NewScript(s, config)
		if err != nil {
			return nil, errors.Wrap(err, "failed to initialize app")
		}
		app.scripts = append(app.scripts, script)
	}
	return app, nil
}

func (app *Application) Start(sync bool) error {
	if sync {
		return app.startSync()
	}
	return app.startAsync()
}

func (app *Application) startAsync() error {
	chs := make([]chan reader.LogEvent, 0)
	for _, s := range app.scripts {
		c := make(chan reader.LogEvent)
		go s.RunAsync(app.ctx, c)
	}
	defer func() {
		for _, c := range chs {
			close(c)
		}
	}()
	f := func(event reader.LogEvent) {
		for _, c := range chs {
			c <- event
		}
	}
	return app.reader.Start(app.ctx, f)
}

func (app *Application) startSync() error {
	f := func(event reader.LogEvent) {
		if event.Err != nil {
			//log
			return
		}
		for _, s := range app.scripts {
			s.Run(app.ctx, event)
		}
	}
	return app.reader.Start(app.ctx, f)
}

// Close closed the application
func (app *Application) Close() error {
	app.cncel()
	return nil
}
