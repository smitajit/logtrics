package graphite

import (
	"fmt"
	"log"
	"net"
	"time"

	graphite "github.com/cyberdelia/go-metrics-graphite"
	"github.com/pkg/errors"
	goMetrics "github.com/rcrowley/go-metrics"
	"github.com/rs/zerolog"
	"github.com/smitajit/logtrics/pkg/config"
	lua "github.com/yuin/gopher-lua"
)

type (
	// Graphite represents the graphite module of the application
	// It store the graphite registry configs and provide method for metrics operations
	Graphite struct {
		registry goMetrics.Registry
		logger   zerolog.Logger
		config   *config.Configuration
	}

	// Counter represents counter metrics
	Counter struct {
		name    string
		counter goMetrics.Counter
	}

	// Timer represents timer metrics
	Timer struct {
		name  string
		timer goMetrics.Timer
	}

	// Meter represents the meter metrics
	Meter struct {
		Name  string
		meter goMetrics.Meter
	}

	// Gauge represents gauge metrics
	Gauge struct {
		name  string
		gauge goMetrics.Gauge
	}
)

// NewGraphite returns a new graphite instance
// It starts the thread which published the metrics in regular interval (config.Graphite.Interval)
func NewGraphite(config *config.Configuration, L *lua.LState, logger zerolog.Logger) (*Graphite, error) {
	var (
		registry = goMetrics.NewRegistry()
		interval = time.Second * time.Duration(config.Graphite.Interval)
		address  = fmt.Sprintf("%s:%d", config.Graphite.Host, config.Graphite.Port)
	)

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, errors.Wrap(err, "graphite connection failed")
	}
	fmt.Println(addr)

	c := graphite.Config{
		Addr:          addr,
		Registry:      registry,
		FlushInterval: time.Duration(config.Graphite.Interval),
		DurationUnit:  time.Second,
		Percentiles:   []float64{0.5, 0.75, 0.95, 0.99, 0.999},
	}

	if config.Graphite.Debug {
		logger.Debug().
			Str("graphite.host", config.Graphite.Host).
			Int("graphite.port", config.Graphite.Port).
			Int("graphite.interval", config.Graphite.Interval).
			Bool("graphite.debug", config.Graphite.Debug).
			Msg("graphite configuration")
		go goMetrics.Log(registry, interval, log.New(logger, "metrics", log.Lmicroseconds))
	}
	go func() {
		for range time.Tick(interval) {
			if err := graphite.Once(c); err != nil {
				logger.Error().Err(err).Msg("failed to send graphite metrics")
			}
		}
	}()
	g := &Graphite{
		config:   config,
		logger:   logger,
		registry: registry,
	}

	return g, nil
}

// LAPICounter is lua binding for counter function on the graphite instance
func (g *Graphite) LAPICounter(L *lua.LState) int {
	metricname := L.ToString(1)
	if metricname == "" {
		L.RaiseError("graphite: invalid counter name")
	}
	c := g.counter(metricname)
	table := L.NewTable()
	L.SetField(table, "inc", L.NewFunction(c.LAPIInc))
	L.SetField(table, "dec", L.NewFunction(c.LAPIDec))
	L.Push(table)
	return 1
}

// LAPIGauge is the lua biding for gauge function call on the graphite instance
func (g *Graphite) LAPIGauge(L *lua.LState) int {
	metricname := L.ToString(1)
	if metricname == "" {
		L.RaiseError("graphite: invalid gauge name")
	}
	m := g.gauge(metricname)
	table := L.NewTable()
	L.SetField(table, "update", L.NewFunction(m.LAPIUpdate))
	L.Push(table)
	return 1
}

// LAPITimer is the lua binding for timer function call on the graphite instance
func (g *Graphite) LAPITimer(L *lua.LState) int {
	metricname := L.ToString(1)
	if metricname == "" {
		L.RaiseError("graphite: invalid timer name")
	}
	m := g.timer(metricname)
	table := L.NewTable()
	L.SetField(table, "update", L.NewFunction(m.LAPIUpdate))
	L.Push(table)
	return 1
}

// LAPIMeter is the lua binding for the meter function call on the graphite instance
func (g *Graphite) LAPIMeter(L *lua.LState) int {
	metricname := L.ToString(1)
	if metricname == "" {
		L.RaiseError("graphite: invalid meter name")
	}
	m := g.meter(metricname)
	table := L.NewTable()
	L.SetField(table, "mark", L.NewFunction(m.LAPIMark))
	L.Push(table)
	return 1
}

// timer return the timer instance for the metrics name
func (g *Graphite) timer(name string) *Timer {
	return &Timer{
		name:  name,
		timer: goMetrics.GetOrRegisterTimer(name, g.registry),
	}
}

// gauge returns the gauge for the metric name
func (g *Graphite) gauge(name string) *Gauge {
	return &Gauge{
		name:  name,
		gauge: goMetrics.GetOrRegisterGauge(name, g.registry),
	}
}

// counter returns the counter for the metrics name
func (g *Graphite) counter(name string) *Counter {
	return &Counter{
		name:    name,
		counter: goMetrics.GetOrRegisterCounter(name, g.registry),
	}
}

// counter returns the counter for the metrics name
func (g *Graphite) meter(name string) *Meter {
	return &Meter{
		Name:  name,
		meter: goMetrics.GetOrRegisterMeter(name, g.registry),
	}
}

// LAPIUpdate is lua binding for update function call on the timer instance
func (t *Timer) LAPIUpdate(L *lua.LState) int {
	i := L.ToInt64(1)
	t.timer.Update(time.Duration(i))
	return 1
}

// LAPIUpdate is lua binding for update function call on the gauge instance
func (g *Gauge) LAPIUpdate(L *lua.LState) int {
	i := L.ToInt64(1)
	g.gauge.Update(i)
	return 1
}

// LAPIMark is the lua binding for mark function call on the meter instance
func (g *Meter) LAPIMark(L *lua.LState) int {
	i := L.ToInt64(1)
	g.meter.Mark(i)
	return 1
}

// LAPIInc is the lua binding for inc function call
func (c *Counter) LAPIInc(L *lua.LState) int {
	i := L.ToInt64(1)
	c.counter.Inc(i)
	return 0
}

// LAPIDec is the lua call back for dec function call
func (c *Counter) LAPIDec(L *lua.LState) int {
	i := L.ToInt64(1)
	c.counter.Dec(i)
	return 0
}
