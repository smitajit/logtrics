package pkg

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/smitajit/logtrics/config"
	"github.com/smitajit/logtrics/pkg/graphite"
	"github.com/smitajit/logtrics/pkg/reader"
	lua "github.com/yuin/gopher-lua"
)

type (
	// Logtric represents the logtrics instance configured in lua
	Logtric struct {
		id       string
		state    *lua.LState
		graphite *graphite.Graphite
		parser   *Parser
		process  *lua.LFunction
	}
)

// NewLogtric returns a new instance of Logtric
func NewLogtric(id string, config *config.Configuration) *Logtric {
	l := &Logtric{
		id: id,
	}
	if config.Graphite != nil {
		l.graphite = &graphite.Graphite{
			Host:     config.Graphite.Host,
			Port:     config.Graphite.Port,
			Interval: config.Graphite.Interval,
		}
	}
	return l
}

// Run runs the Logtric instance
func (l *Logtric) Run(ctx context.Context, event reader.LogEvent) error {
	p := lua.P{
		Fn:      l.process,
		NRet:    0,
		Protect: true,
	}

	args := []string{event.Source, event.Line}
	substrings, ok := l.parser.FindSubStrings(event.Line)
	if !ok {
		// TODO log maybe?
		fmt.Println("doesn't match")
		return nil
	}
	args = append(args, substrings...)
	lParams := make([]lua.LValue, 0)
	for _, p := range args {
		lParams = append(lParams, lua.LString(p))
	}
	l.state.SetGlobal("graphite", l.state.NewFunction(l.graphite.LGraphite))
	err := l.state.CallByParam(p, lParams...)
	if err != nil && err.Error() != "nil" {
		return err
	}
	return nil
}

func (l *Logtric) updateProcess(v lua.LValue) error {
	p, ok := v.(*lua.LFunction)
	if !ok {
		return fmt.Errorf("invalid process config")
	}
	l.process = p
	return nil
}

func (l *Logtric) updateExpression(v lua.LValue) error {
	if nil == v {
		return fmt.Errorf("expression is required") //TODO better error
	}
	expr, ok := v.(lua.LString)
	if !ok {
		return fmt.Errorf("invalid lua configuration") // TODO better error
	}
	p, err := NewParser(expr.String())
	if err != nil {
		return errors.Wrap(err, "invalid lua config")
	}
	l.parser = p
	return nil
}
