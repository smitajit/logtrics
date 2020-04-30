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
	config  *config.Configuration
	logger  zerolog.Logger
}

//NewApplication returns a new Application instance
func NewApplication(config *config.Configuration, reader reader.LogReader, scripts ...string) (*Application, error) {
	app := &Application{
		reader:  reader,
		scripts: make([]*Script, 0),
		config:  config,
		logger:  config.Logger("application"),
	}
	for _, s := range scripts {
		script, err := NewScript(s, config)
		if err != nil {
			return nil, errors.Wrap(err, "failed to initialize app")
		}
		app.scripts = append(app.scripts, script)
	}
	return app, nil
}

// RunAsync runs the application
// returns error in case of any failure
// parameter async represents the mode of application. If set as false all the scrips will run in single go routine, otherwise each script will run in its own go routine
// note:  this is a blocking call.
func (app *Application) RunAsync(ctx context.Context) error {
	chs := make([]chan reader.LogEvent, 0)
	for _, s := range app.scripts {
		c := make(chan reader.LogEvent, app.config.BufferSize)
		go s.RunAsync(ctx, c)
		chs = append(chs, c)
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
	return app.reader.Start(ctx, f)
}

// Run runs the application
// returns error in case of any failure
// note:  this is a blocking call.
func (app *Application) Run(ctx context.Context) error {
	f := func(event reader.LogEvent) {
		if event.Err != nil {
			//log
			return
		}
		for _, s := range app.scripts {
			s.Run(ctx, event)
		}
	}
	return app.reader.Start(ctx, f)
}
