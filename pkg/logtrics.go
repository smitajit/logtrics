package pkg

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/smitajit/logtrics/config"
	"github.com/smitajit/logtrics/pkg/graphite"
	"github.com/smitajit/logtrics/pkg/reader"
	lua "github.com/yuin/gopher-lua"
)

type (
	// Logtric represents the logtrics instance configured in lua
	// it stores the lua script states and provides runtime bindings to lua
	Logtric struct {
		id       string
		state    *lua.LState
		parser   *Parser
		process  *lua.LFunction
		config   *config.Configuration
		graphite *graphite.Graphite
		logger   zerolog.Logger
	}
)

// NewLogtric returns a new instance of Logtric
func NewLogtric(id string, config *config.Configuration, state *lua.LState, v lua.LValue) (*Logtric, error) {
	table, ok := v.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("invalid logtric config")
	}

	p := table.RawGet(lua.LString("process"))
	process, ok := p.(*lua.LFunction)
	if !ok || process == nil {
		return nil, fmt.Errorf("invalid logtric config. Process config error")
	}

	config, err := config.Merge(table)
	if err != nil {
		return nil, err
	}

	parser, err := NewParser(config.Expression)
	if err != nil {
		return nil, err //TODO wrap error may be
	}

	l := &Logtric{
		id:      id,
		state:   state,
		config:  config,
		process: process,
		parser:  parser,
		logger:  config.Logger(id),
	}
	return l, nil
}

// Run runs the Logtric instance
func (l *Logtric) Run(ctx context.Context, event reader.LogEvent) error {
	p := lua.P{
		Fn:      l.process,
		NRet:    0,
		Protect: true,
	}

	l.state.SetGlobal("debug", l.state.NewFunction(l.LAPIDebug))
	l.state.SetGlobal("error", l.state.NewFunction(l.LAPIError))
	l.state.SetGlobal("info", l.state.NewFunction(l.LAPIInfo))
	l.state.SetGlobal("graphite", l.state.NewFunction(l.LAPIGraphite))

	args := []string{event.Source, event.Line}
	substrings, ok := l.parser.FindSubStrings(event.Line)
	if !ok {
		l.logger.Debug().Msg("expression doesn't match")
		return nil
	}
	args = append(args, substrings...)
	lParams := make([]lua.LValue, 0)
	for _, p := range args {
		lParams = append(lParams, lua.LString(p))
	}
	err := l.state.CallByParam(p, lParams...)
	if err != nil && err.Error() != "nil" {
		return err
	}
	return nil
}

// LAPIInfo represents the lua binding for info() function call
func (l *Logtric) LAPIInfo(L *lua.LState) int {
	n := L.GetTop()
	if n < 1 {
		L.RaiseError("parameter required for info")
	}
	args := make([]interface{}, 0)
	for i := 2; i <= n; i++ {
		v := L.Get(i)
		args = append(args, v.String())
	}
	l.logger.Info().Msgf(L.ToString(1), args...)
	return 0
}

// LAPIDebug represents the lua binding for debug() function call
func (l *Logtric) LAPIDebug(L *lua.LState) int {
	n := L.GetTop()
	if n < 1 {
		L.RaiseError("parameter required for debug")
	}
	args := make([]interface{}, 0)
	for i := 2; i <= n; i++ {
		v := L.Get(i)
		args = append(args, v.String())
	}
	l.logger.Debug().Msgf(L.ToString(1), args...)
	return 0
}

// LAPIError represent the lua binding for error() function call
func (l *Logtric) LAPIError(L *lua.LState) int {
	n := L.GetTop()
	if n < 1 {
		L.RaiseError("parameter required for error")
	}
	args := make([]interface{}, 0)
	for i := 2; i <= n; i++ {
		v := L.Get(i)
		args = append(args, v.String())
	}
	l.logger.Error().Msgf(L.ToString(1), args...)
	return 0
}

// LAPIGraphite is represents the lua binding for graphite() api call
func (l *Logtric) LAPIGraphite(L *lua.LState) int {
	if l.graphite == nil {
		g, err := graphite.NewGraphite(l.config, L, l.logger)
		if err != nil {
			L.RaiseError(err.Error())
		}
		l.graphite = g
	}
	table := L.NewTable()
	L.SetField(table, "counter", L.NewFunction(l.graphite.LAPICounter))
	L.SetField(table, "timer", L.NewFunction(l.graphite.LAPICounter))
	L.SetField(table, "gauge", L.NewFunction(l.graphite.LAPICounter))
	L.Push(table)
	return 1
}
