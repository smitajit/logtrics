package pkg

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/smitajit/logtrics/config"
	"github.com/smitajit/logtrics/pkg/reader"
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
	// defer state.Close()
	state.SetGlobal("debug", state.NewFunction(s.lDebug))
	state.SetGlobal("logtrics", state.NewFunction(s.lCompile))
	if err := state.DoFile(s.Path); err != nil {
		return nil, err
	}
	return s, nil
}

// RunAsync ...
func (s *Script) RunAsync(ctx context.Context, c <-chan reader.LogEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-c:
			s.Run(ctx, event)
		}
	}
}

// Run runs the script
func (s *Script) Run(ctx context.Context, event reader.LogEvent) {
	for _, l := range s.logtrics {
		l.Run(ctx, event)
	}
}

func (s *Script) lDebug(L *lua.LState) int {
	n := L.GetTop()
	if n < 1 {
		L.ArgError(1, "debug arguments required")
	}
	args := make([]interface{}, 0)
	for i := 2; i <= n; i++ {
		v := L.Get(i)
		args = append(args, v.String())
	}
	s.logger.Debug().Msgf(L.ToString(1), args...)
	return 0
}

// lCompile is lua script compilation callback
func (s *Script) lCompile(L *lua.LState) int {
	l := NewLogtric(fmt.Sprintf("%s:%d", s.Path, len(s.logtrics)), s.config)
	// parsing the lua script
	table := L.ToTable(1)
	var err error = nil
	table.ForEach(func(k lua.LValue, v lua.LValue) {
		// checking if previous parsing caused any error
		if err != nil {
			return
		}
		switch k {
		case lua.LString("graphite"):
			err = l.graphite.Update(v)
		case lua.LString("expression"):
			err = l.updateExpression(v)
		case lua.LString("process"):
			err = l.updateProcess(v)
		}
	})
	if err != nil {
		L.ArgError(1, errors.Wrap(err, "compilation failed").Error())
	}
	l.state = L
	s.logtrics = append(s.logtrics, l)
	return 1
}
