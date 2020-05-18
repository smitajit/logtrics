package pkg

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/smitajit/logtrics/pkg/config"
	"github.com/smitajit/logtrics/pkg/graphite"
	"github.com/smitajit/logtrics/pkg/reader"
	lua "github.com/yuin/gopher-lua"
)

type (
	// Logtric represents the logtrics instance configured in lua
	// it stores the lua script states and provides runtime bindings to lua
	Logtric struct {
		name      string
		state     *lua.LState
		parser    Parser
		handler   *lua.LFunction
		config    *config.Configuration
		graphite  *graphite.Graphite
		logger    zerolog.Logger
		scheduler *Scheduler
	}

	// Scheduler represents the scheduler config inside logtrics
	Scheduler struct {
		interval int
		handler  *lua.LFunction
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
		return nil, fmt.Errorf("handler not found")
	}

	s := table.RawGet(lua.LString("scheduler"))
	scheduler, err := scheduler(s)
	if err != nil {
		return nil, err
	}

	config, err = config.Merge(table)
	if err != nil {
		return nil, err
	}

	l := &Logtric{
		name:      name,
		state:     state,
		config:    config,
		handler:   handler,
		parser:    parser,
		scheduler: scheduler,
		logger:    config.Logger(fmt.Sprintf("%s:[%s]", script, name)),
	}

	l.bindApis()

	if l.scheduler != nil {
		l.scheduler.Start()
	}
	return l, nil
}

func (l *Logtric) bindApis() {
	// binding logging apis
	l.state.SetGlobal("fatal", l.state.NewFunction(l.LAPIFatal))
	l.state.SetGlobal("error", l.state.NewFunction(l.LAPIError))
	l.state.SetGlobal("warn", l.state.NewFunction(l.LAPIWarn))
	l.state.SetGlobal("info", l.state.NewFunction(l.LAPIInfo))
	l.state.SetGlobal("debug", l.state.NewFunction(l.LAPIDebug))
	l.state.SetGlobal("trace", l.state.NewFunction(l.LAPITrace))

	// binding graphite api
	l.state.SetGlobal("graphite", l.state.NewFunction(l.LAPIGraphite))
}

// Start ...
func (s *Scheduler) Start() {
	ticker := time.NewTicker(time.Second * time.Duration(s.interval))
	go func() {
		for range ticker.C {
			state := lua.NewState()
			p := lua.P{
				Fn:      s.handler,
				NRet:    0,
				Protect: true,
			}
			if err := state.CallByParam(p); err != nil {
				fmt.Println(err)
			}
		}
	}()
}

// TODO(smitajit) handle it properly
func scheduler(v lua.LValue) (*Scheduler, error) {
	table, ok := v.(*lua.LTable)
	if !ok {
		return nil, errors.New("invalid scheduler config")
	}

	i := table.RawGet(lua.LString("interval"))
	in, ok := i.(lua.LNumber)
	if !ok {
		return nil, errors.New("invalid scheduler config, interval is wrong")
	}
	interval, err := strconv.Atoi(in.String())
	if err != nil {
		return nil, errors.New("invalid scheduler config, interval is wrong1")
	}

	h := table.RawGet(lua.LString("handler"))
	handler, ok := h.(*lua.LFunction)
	if !ok {
		return nil, errors.New("invalid scheduler config, handler is wrong")
	}
	return &Scheduler{
		interval: interval,
		handler:  handler,
	}, nil
}

// Run runs the Logtric instance
func (l *Logtric) Run(ctx context.Context, event reader.LogEvent) error {
	p := lua.P{
		Fn:      l.handler,
		NRet:    0,
		Protect: true,
	}

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

func (l *Logtric) parseLogArgs(name string, state *lua.LState) (msg string, args []interface{}) {
	top := state.GetTop()
	if top < 1 {
		state.RaiseError("parameter required for " + name)
	}
	msg = state.ToString(1)
	for i := 2; i <= top; i++ {
		v := state.Get(i)
		t, ok := v.(*lua.LTable)
		if ok {
			str := "["
			t.ForEach(func(k, v lua.LValue) {
				str += fmt.Sprintf(" %s = %s, ", k.String(), v.String())
			})
			str += "]"
			args = append(args, str)
		} else {
			args = append(args, v.String())
		}
	}
	return
}

// LAPIInfo represents the lua binding for info() function call
func (l *Logtric) LAPIInfo(state *lua.LState) int {
	msg, args := l.parseLogArgs("info", state)
	l.logger.Info().Msgf(msg, args...)
	return 0
}

// LAPIDebug represents the lua binding for debug() function call
func (l *Logtric) LAPIDebug(state *lua.LState) int {
	msg, args := l.parseLogArgs("debug", state)
	l.logger.Debug().Msgf(msg, args...)
	return 0
}

// LAPIWarn represent the lua binding for error() function call
func (l *Logtric) LAPIWarn(state *lua.LState) int {
	msg, args := l.parseLogArgs("warn", state)
	l.logger.Warn().Msgf(msg, args...)
	return 0
}

// LAPIError represent the lua binding for error() function call
func (l *Logtric) LAPIError(state *lua.LState) int {
	msg, args := l.parseLogArgs("error", state)
	l.logger.Error().Msgf(msg, args...)
	return 0
}

// LAPIFatal represent the lua binding for fatal() function call
func (l *Logtric) LAPIFatal(state *lua.LState) int {
	msg, args := l.parseLogArgs("fatal", state)
	l.logger.Fatal().Msgf(msg, args...)
	return 0
}

// LAPITrace represent the lua binding for fatal() function call
func (l *Logtric) LAPITrace(state *lua.LState) int {
	msg, args := l.parseLogArgs("trace", state)
	l.logger.Trace().Msgf(msg, args...)
	return 0
}

// LAPIGraphite is represents the lua binding for graphite() api call
func (l *Logtric) LAPIGraphite(state *lua.LState) int {
	if l.graphite == nil {
		g, err := graphite.NewGraphite(l.config, state, l.logger)
		if err != nil {
			state.RaiseError(err.Error())
		}
		l.graphite = g
	}
	table := state.NewTable()
	state.SetField(table, "counter", state.NewFunction(l.graphite.LAPICounter))
	state.SetField(table, "timer", state.NewFunction(l.graphite.LAPITimer))
	state.SetField(table, "gauge", state.NewFunction(l.graphite.LAPIGauge))
	state.SetField(table, "meter", state.NewFunction(l.graphite.LAPIMeter))
	state.Push(table)
	return 1
}
