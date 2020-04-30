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
		name     string
		state    *lua.LState
		parser   Parser
		handler  *lua.LFunction
		config   *config.Configuration
		graphite *graphite.Graphite
		logger   zerolog.Logger
		EmitCh   chan reader.LogEvent
	}
)

// NewLogtric returns a new instance of Logtric
func NewLogtric(script string, config *config.Configuration, state *lua.LState, table *lua.LTable) (*Logtric, error) {
	name := table.RawGet(lua.LString("name")).String()
	if name == "" || name == "nil" {
		name = "?"
	}

	p := table.RawGet(lua.LString("parser"))
	parserTable, ok := p.(*lua.LTable)
	if !ok || parserTable == nil {
		return nil, fmt.Errorf("parser not found")
	}
	parser, err := NewParser(parserTable)
	if err != nil {
		return nil, err //TODO wrap error may be
	}

	h := table.RawGet(lua.LString("handler"))
	handler, ok := h.(*lua.LFunction)
	if !ok || handler == nil {
		return nil, fmt.Errorf("process not found")
	}

	config, err = config.Merge(table)
	if err != nil {
		return nil, err
	}

	l := &Logtric{
		name:    name,
		state:   state,
		config:  config,
		handler: handler,
		parser:  parser,
		logger:  config.Logger(fmt.Sprintf("%s:[%s]", script, name)),
	}
	return l, nil
}

// Run runs the Logtric instance
func (l *Logtric) Run(ctx context.Context, event reader.LogEvent) error {
	p := lua.P{
		Fn:      l.handler,
		NRet:    0,
		Protect: true,
	}
	// binding logging apis
	l.state.SetGlobal("fatal", l.state.NewFunction(l.LAPIFatal))
	l.state.SetGlobal("error", l.state.NewFunction(l.LAPIError))
	l.state.SetGlobal("warn", l.state.NewFunction(l.LAPIWarn))
	l.state.SetGlobal("info", l.state.NewFunction(l.LAPIInfo))
	l.state.SetGlobal("debug", l.state.NewFunction(l.LAPIDebug))
	l.state.SetGlobal("trace", l.state.NewFunction(l.LAPITrace))

	// bindings for emit api
	l.state.SetGlobal("emit", l.state.NewFunction(l.LAPIEmit))

	// binding graphite api
	l.state.SetGlobal("graphite", l.state.NewFunction(l.LAPIGraphite))

	// args := []string{event.Source, event.Line}
	substrings, ok := l.parser.FindSubStrings(event.Line)
	if !ok {
		l.logger.Debug().Msg("expression doesn't match")
		return nil
	}

	table := l.state.NewTable()
	table.RawSetString("_source", lua.LString(event.Source))
	table.RawSetString("_line", lua.LString(event.Line))

	for k, v := range substrings {
		table.RawSetString(k, lua.LString(v))
	}
	err := l.state.CallByParam(p, table)
	if err != nil && err.Error() != "nil" {
		return err
	}
	return nil
}

// LAPIEmit represents the lua binding for emit api call
func (l *Logtric) LAPIEmit(L *lua.LState) int {
	log := L.ToString(1)
	l.EmitCh <- reader.LogEvent{Source: l.name, Line: log, Err: nil}
	return 1
}

func (l *Logtric) parseLogArgs(name string, L *lua.LState) (msg string, args []interface{}) {
	top := L.GetTop()
	if top < 1 {
		L.RaiseError("parameter required for " + name)
	}
	msg = L.ToString(1)
	for i := 2; i <= top; i++ {
		v := L.Get(i)
		t, ok := v.(*lua.LTable)
		if ok {
			str := "["
			t.ForEach(func(k, v lua.LValue) {
				str = str + fmt.Sprintf(" %s = %s, ", k.String(), v.String())
			})
			str = str + "]"
			args = append(args, str)
		} else {
			args = append(args, v.String())
		}
	}
	return
}

// LAPIInfo represents the lua binding for info() function call
func (l *Logtric) LAPIInfo(L *lua.LState) int {
	msg, args := l.parseLogArgs("info", L)
	l.logger.Info().Msgf(msg, args...)
	return 0
}

// LAPIDebug represents the lua binding for debug() function call
func (l *Logtric) LAPIDebug(L *lua.LState) int {
	msg, args := l.parseLogArgs("debug", L)
	l.logger.Debug().Msgf(msg, args...)
	return 0
}

// LAPIWarn represent the lua binding for error() function call
func (l *Logtric) LAPIWarn(L *lua.LState) int {
	msg, args := l.parseLogArgs("warn", L)
	l.logger.Warn().Msgf(msg, args...)
	return 0
}

// LAPIError represent the lua binding for error() function call
func (l *Logtric) LAPIError(L *lua.LState) int {
	msg, args := l.parseLogArgs("error", L)
	l.logger.Error().Msgf(msg, args...)
	return 0
}

// LAPIFatal represent the lua binding for fatal() function call
func (l *Logtric) LAPIFatal(L *lua.LState) int {
	msg, args := l.parseLogArgs("fatal", L)
	l.logger.Fatal().Msgf(msg, args...)
	return 0
}

// LAPITrace represent the lua binding for fatal() function call
func (l *Logtric) LAPITrace(L *lua.LState) int {
	msg, args := l.parseLogArgs("trace", L)
	l.logger.Trace().Msgf(msg, args...)
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
	L.SetField(table, "timer", L.NewFunction(l.graphite.LAPITimer))
	L.SetField(table, "gauge", L.NewFunction(l.graphite.LAPIGauge))
	L.SetField(table, "meter", L.NewFunction(l.graphite.LAPIMeter))
	L.Push(table)
	return 1
}
