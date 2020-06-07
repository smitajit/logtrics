package logtrics

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/smitajit/logtrics/config"
	"github.com/smitajit/logtrics/reader"
)

// Application represents this application
// it stores all the application states and maintains the runtime
type Application struct {
	readers []reader.LogReader
	scripts []*Script
	conf    *config.Configuration
	logger  zerolog.Logger
}

//NewApplication returns a new Application instance
func NewApplication(conf *config.Configuration, readers ...reader.LogReader) (*Application, error) {
	app := &Application{
		readers: readers,
		scripts: make([]*Script, 0),
		conf:    conf,
		logger:  conf.Logger("application"),
	}

	files, err := scripts(conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get script files")
	}
	for _, f := range files {
		script, err := NewScript(f, conf)
		if err != nil {
			return nil, errors.Wrap(err, "failed to initialize app")
		}
		app.scripts = append(app.scripts, script)
	}
	return app, nil
}

// RunAsync runs the application in async mode.
// returns error in case of any failure
// note:  this is a blocking call.
func (app *Application) RunAsync(ctx context.Context) error {
	chs := make([]chan reader.LogEvent, 0)
	for _, s := range app.scripts {
		c := make(chan reader.LogEvent, app.conf.BufferSize)
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
	return app.run(ctx, f)
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
	return app.run(ctx, f)
}

func (app *Application) run(ctx context.Context, fn func(event reader.LogEvent)) error {
	for _, reader := range app.readers {
		if err := reader.Start(ctx, fn); err != nil {
			return errors.Wrap(err, "failed to start the readers")
		}
	}
	return nil
}

func scripts(conf *config.Configuration) ([]string, error) {
	if conf.ScriptFile != "" {
		return []string{conf.ScriptFile}, nil
	}
	scripts := make([]string, 0)
	err := filepath.Walk(conf.ScriptDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".lua") {
			scripts = append(scripts, path)
		}
		return nil
	})
	if len(scripts) == 0 {
		return nil, fmt.Errorf("no scripts found")
	}
	return scripts, err
}
