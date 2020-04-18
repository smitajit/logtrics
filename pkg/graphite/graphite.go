package graphite

import (
	"fmt"
	"strconv"

	lua "github.com/yuin/gopher-lua"
)

type (
	// Graphite handles the graphite
	Graphite struct {
		Host     string
		Port     int
		Interval int
	}

	// Counter represents counter metrics
	Counter struct {
		Name string
	}
	// Timer represents timer metrics
	Timer struct {
		Name string
	}

	// Gauge represents gauge metrics
	Gauge struct {
		Name string
	}
)

// Update updates self with lua table
func (g *Graphite) Update(v lua.LValue) error {
	if v == nil {
		// default values are already present. No need to panic :)
		return nil
	}
	table, ok := v.(*lua.LTable)
	if !ok {
		return fmt.Errorf("invalid graphite config")
	}
	var err error = nil
	table.ForEach(func(k, v lua.LValue) {
		if err != nil {
			return
		}
		switch k {
		case lua.LString("host"):
			g.Host = v.String()
			if g.Host == "" {
				err = fmt.Errorf("invalid graphite config")
			}
		case lua.LString("port"):
			g.Port, err = strconv.Atoi(v.String())
		case lua.LString("interval"):
			g.Port, err = strconv.Atoi(v.String())
		default:
			err = fmt.Errorf("invalid graphite config [%s]", k.String())
		}
	})
	return err
}

// LGraphite is the lua binding for graphite() function
func (g *Graphite) LGraphite(L *lua.LState) int {
	table := L.NewTable()
	L.SetField(table, "counter", L.NewFunction(g.LCounter))
	L.Push(table)
	return 1
}

// LCounter is lua binding for counter function on graphite
func (g *Graphite) LCounter(L *lua.LState) int {
	metricname := L.ToString(1)
	if metricname == "" {
		L.ArgError(1, "no metric name found for the counter")
	}
	c := g.counter(metricname)
	table := L.NewTable()
	L.SetField(table, "inc", L.NewFunction(c.LInc))
	L.SetField(table, "dec", L.NewFunction(c.LDec))
	L.Push(table)
	return 1
}

// LGauge is the lua biding for gauge function call on graphite
func (g *Graphite) LGauge(L *lua.LState) int {
	metricname := L.ToString(1)
	if metricname == "" {
		L.ArgError(1, "no metric name found for the counter")
	}
	m := g.gauge(metricname)
	table := L.NewTable()
	L.SetField(table, "update", L.NewFunction(m.LUpdate))
	L.Push(table)
	return 1
}

// LTimer is the lua binding for timer function call n graphite
func (g *Graphite) LTimer(L *lua.LState) int {
	metricname := L.ToString(1)
	if metricname == "" {
		L.ArgError(1, "no metric name found for the counter")
	}
	m := g.timer(metricname)
	table := L.NewTable()
	L.SetField(table, "update", L.NewFunction(m.LUpdate))
	L.Push(table)
	return 1
}

// timer return the timer instance for the metrics name
func (g *Graphite) timer(name string) *Timer {
	return &Timer{
		Name: name,
	}
}

// gauge returns the gauge for metric name
func (g *Graphite) gauge(name string) *Gauge {
	return &Gauge{
		Name: name,
	}
}

// counter returns the counter for the metrics name
func (g *Graphite) counter(name string) *Counter {
	return &Counter{
		Name: name,
	}
}

// LUpdate is lua binding for update function call on timer
func (t *Timer) LUpdate(L *lua.LState) int {
	i := L.ToNumber(1)
	t.update(float64(i))
	return 1
}

// LUpdate is lua binding for update function call on gauge
func (g *Gauge) LUpdate(L *lua.LState) int {
	i := L.ToNumber(1)
	g.update(float64(i))
	return 1
}

// update updates the timer value
func (t *Timer) update(i float64) {
	fmt.Printf("updating the timer [%s] with value [%f]\n", t.Name, i)
}

// update updates gauge values
func (g *Gauge) update(i float64) {
	fmt.Printf("updating the gauge [%s] with value [%f]\n", g.Name, i)
}

// LInc is the lua binding for int() call
func (c *Counter) LInc(L *lua.LState) int {
	i := L.ToInt(1)
	c.inc(i)
	return 0
}

// inc increments the counter value
func (c *Counter) inc(i int) {
	fmt.Printf("incrementing counter [%s] with value [%d]\n", c.Name, i)
}

// LDec is the lua call back for dec function call
func (c *Counter) LDec(L *lua.LState) int {
	i := L.ToInt(1)
	c.inc(i)
	return 0
}

// dec decrements the counter value
func (c *Counter) dec(i int) {
	fmt.Printf("incrementing counter [%s] with value [%d]\n", c.Name, i)
}
