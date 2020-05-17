package pkg

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/smitajit/logtrics/pkg/config"
	lua "github.com/yuin/gopher-lua"
)

type (
	// Script represents the logtrics lua script
	Script struct {
		Path     string
		logtrics []*Logtric
		config   *config.Configuration
		logger   zerolog.Logger
	}
)

// NewScript returns a new Script instance which represents a lua script file
func NewScript(path string, config *config.Configuration) (*Script, error) {
	s := &Script{
		Path:     path,
		config:   config,
		logtrics: make([]*Logtric, 0),
		logger:   config.Logger(path),
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
func (s *Script) RunAsync(ctx context.Context, c <-chan LogEvent) {
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
func (s *Script) Run(ctx context.Context, event LogEvent) {
	logger := s.config.Logger(s.Path)
	logger.Debug().Msgf("executing script")
	for _, l := range s.logtrics {
		if err := l.Run(ctx, event); err != nil {
			logger.Error().Err(err).Msgf("script execution error")
		}
	}
}

// LAPILogtric represents lua binding for logtric initialization
func (s *Script) LAPILogtric(L *lua.LState) int {
	// parsing the lua script
	table := L.ToTable(1)
	l, err := NewLogtric(s.Path, s.config, L, table)
	if err != nil {
		L.RaiseError(err.Error())
	}
	s.logtrics = append(s.logtrics, l)
	return 1
}
