package logtrics

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/smitajit/logtrics/config"
	"github.com/smitajit/logtrics/reader"
	lua "github.com/yuin/gopher-lua"
)

type (
	// Script represents the logtrics lua script
	Script struct {
		Path     string
		logtrics []*Logtric
		conf     *config.Configuration
		logger   zerolog.Logger
	}
)

// NewScript returns a new Script instance which represents a lua script file
func NewScript(path string, conf *config.Configuration) (*Script, error) {
	s := &Script{
		Path:     path,
		conf:     conf,
		logtrics: make([]*Logtric, 0),
		logger:   conf.Logger(path),
	}
	state := lua.NewState()
	state.SetGlobal("logtrics", state.NewFunction(s.LAPILogtric))
	if err := state.DoFile(s.Path); err != nil {
		return nil, err
	}
	return s, nil
}

// RunAsync runs the script in async mode
// It consumes the log event from the channel
// note: this is a blocking call
func (s *Script) RunAsync(ctx context.Context, c <-chan reader.LogEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-c:
			s.logger.Debug().Msgf("log event received from reader : %s", event.Source)
			s.Run(ctx, event)
		}
	}
}

// Run runs the script
// This is non blocking call
func (s *Script) Run(ctx context.Context, event reader.LogEvent) {
	logger := s.conf.Logger(s.Path)
	logger.Debug().Msgf("executing script")
	for _, l := range s.logtrics {
		if err := l.Run(ctx, event); err != nil {
			logger.Error().Err(err).Msgf("script execution error")
		}
	}
}

// LAPILogtric represents lua binding for logtric initialization
func (s *Script) LAPILogtric(state *lua.LState) int {
	// parsing the lua script
	table := state.ToTable(1)
	l, err := NewLogtric(s.Path, s.conf, state, table)
	if err != nil {
		state.RaiseError(err.Error())
	}
	s.logtrics = append(s.logtrics, l)
	return 1
}
