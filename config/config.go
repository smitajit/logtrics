package config

import (
	"fmt"
	"log/syslog"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/jinzhu/copier"
	"github.com/rs/zerolog"
	lua "github.com/yuin/gopher-lua"
)

var (
	//nolint:gochecknoglobals
	lvlMap = map[string]zerolog.Level{
		"debug": zerolog.DebugLevel, "info": zerolog.InfoLevel,
		"warn": zerolog.WarnLevel, "error": zerolog.ErrorLevel, "": zerolog.Disabled,
	}
)

type (
	// Configuration represents the application's configuration
	Configuration struct {
		Mode       string    `toml:"mode"`
		Expression string    `toml:"expression"`
		ScriptFile string    `toml:"scriptfile"`
		ScriptDir  string    `toml:"scriptdir"`
		BufferSize int       `toml:"buffersize"`
		Graphite   *Graphite `toml:"graphite"`
		UDP        *UDP      `toml:"udp"`
		TCP        *TCP      `toml:"tcp"`
		Logging    *Logging  `toml:"logging"`
	}

	// UDP configuration
	UDP struct {
		Host string `toml:"host"`
		Port int    `toml:"port"`
	}

	// TCP configuration
	TCP struct {
		Host string `toml:"host"`
		Port int    `toml:"port"`
	}

	// Logging configuration
	Logging struct {
		Type  string `toml:"type"`
		Level string `toml:"level"`
	}

	// Graphite configuration
	Graphite struct {
		Host     string `toml:"host"`
		Port     int    `toml:"port"`
		Interval int    `toml:"interval"`
		Debug    bool   `toml:"debug"`
	}
)

// Merge returns new Configuration after merging the values from lua
func (c *Configuration) Merge(table *lua.LTable) (*Configuration, error) {
	var (
		merged       = new(Configuration)
		err    error = nil
	)
	if err := copier.Copy(merged, c); err != nil {
		return nil, err //TODO wrap may be?
	}
	table.ForEach(func(k lua.LValue, v lua.LValue) {
		if err != nil {
			return
		}
		switch k {
		case lua.LString("process"), lua.LString("timer"), lua.LString("name"):
			//ignore
		case lua.LString("graphite"):
			if merged.Graphite == nil {
				merged.Graphite = &Graphite{}
			}
			err = merged.Graphite.Merge(v)
		case lua.LString("logging"):
			if merged.Logging == nil {
				merged.Logging = &Logging{}
			}
			err = merged.Logging.Merge(v)
		case lua.LString("expression"):
			merged.Expression = v.String()
		case lua.LString("sctriptfile"), lua.LString("scriptdir"), lua.LString("mode"), lua.LString("tcp"), lua.LString("udp"):
			err = fmt.Errorf("modification is not supported for [%s]", k.String())
		default:
			err = fmt.Errorf("invalid key %s", k.String())
		}
	})
	return merged, err
}

// Merge ...
func (l *Logging) Merge(v lua.LValue) error {
	table, ok := v.(*lua.LTable)
	if !ok {
		return fmt.Errorf("invalid logging configuration")
	}
	var err error = nil
	table.ForEach(func(k, v lua.LValue) {
		if err != nil {
			return
		}
		switch k {
		case lua.LString("type"):
			l.Type = v.String()
		case lua.LString("level"):
			l.Level = v.String()
		default:
			err = fmt.Errorf("invalid logging config")
			return
		}
	})
	return err
}

// Merge ...
func (g *Graphite) Merge(v lua.LValue) error {
	table, ok := v.(*lua.LTable)
	if !ok {
		return fmt.Errorf("invalid graphite configuration")
	}
	var err error = nil
	table.ForEach(func(k, v lua.LValue) {
		if err != nil {
			return
		}
		switch k {
		case lua.LString("host"):
			g.Host = v.String()
		case lua.LString("port"):
			g.Port, err = strconv.Atoi(v.String())
			if err != nil {
				return
			}
		case lua.LString("interval"):
			g.Interval, err = strconv.Atoi(v.String())
			if err != nil {
				return
			}
		case lua.LString("debug"):
			g.Debug, err = strconv.ParseBool(v.String())
			if err != nil {
				return
			}
		}
	})
	return err
}

// Logger returns the application logger
func (c *Configuration) Logger(source string) (logger zerolog.Logger) {
	level := lvlMap[c.Logging.Level]
	switch {
	case c.Logging.Type == "syslog":
		writer, err := syslog.New(0, path.Base(os.Args[0]))
		if err != nil {
			panic(err)
		}
		logger = zerolog.New(zerolog.SyslogLevelWriter(writer))
	default:
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	}
	return logger.With().Timestamp().Str("source", source).Logger().Level(level)
}
