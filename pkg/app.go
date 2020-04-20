package pkg

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/smitajit/logtrics/config"
	"github.com/smitajit/logtrics/pkg/reader"
)

// Application represents this application
// it stores all the application states and maintains the runtime
type Application struct {
	reader  reader.LogReader
	scripts []*Script
	ctx     context.Context
	cncel   context.CancelFunc
	config  *config.Configuration
	logger  zerolog.Logger
}

//NewApp returns a new Application instance
func NewApp(config *config.Configuration, reader reader.LogReader, scripts ...string) (*Application, error) {
	app := &Application{
		reader:  reader,
		scripts: make([]*Script, 0),
		config:  config,
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

// Start starts the application
// returns error in case of any failure
// parameter sync represents the mode of application. If set as false all the scrips will run in single go routine, otherwise each script will run in its own go routine
// note:  this is a blocking call.
func (app *Application) Start(sync bool) error {
	defer func() { _ = app.Stop() }()
	if sync {
		return app.startSync()
	}
	return app.startAsync()
}

func (app *Application) startAsync() error {
	chs := make([]chan reader.LogEvent, app.config.BufferSize)
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

// Stop closed the application
func (app *Application) Stop() error {
	app.cncel()
	return nil
}
