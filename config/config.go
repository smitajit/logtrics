package config

import (
	"fmt"
	"log/syslog"
	"os"
	"path"
	"strings"

	"github.com/rs/zerolog"
)

var (
	lvlMap = map[string]zerolog.Level{
		"debug": zerolog.DebugLevel, "info": zerolog.InfoLevel,
		"warn": zerolog.WarnLevel, "error": zerolog.ErrorLevel, "": zerolog.Disabled,
	}
)

type (
	// Configuration represents the configuration of the service
	Configuration struct {
		Mode       string    `toml:"mode"`
		ScriptFile string    `toml:"scriptfile"`
		ScriptDir  string    `toml:"scriptdir"`
		Graphite   *Graphite `toml:"graphite"`
		Logging    *Logging  `toml:"logging"`
	}

	Logging struct {
		Type  string `toml:"type"`
		Level string `toml:"level"`
	}

	// Graphite represents the application's graphite configuration
	Graphite struct {
		Host     string `toml:"host"`
		Port     int    `toml:"port"`
		Interval int    `toml:"interval"`
	}
)

func (c *Configuration) Logger(unit string) (logger zerolog.Logger) {
	level := lvlMap[c.Logging.Level]
	switch {
	case c.Logging.Type == "syslog":
		writer, err := syslog.New(0, path.Base(os.Args[0]))
		if err != nil {
			panic(err)
		}
		logger = zerolog.New(zerolog.SyslogLevelWriter(writer))
	default:
		out := zerolog.ConsoleWriter{Out: os.Stdout, NoColor: false}
		out.FormatLevel = func(i interface{}) string { return strings.ToUpper(fmt.Sprintf("%s", i)) }
		out.FormatFieldName = func(i interface{}) string { return fmt.Sprintf("%s:", i) }
		out.FormatFieldValue = func(i interface{}) string { return fmt.Sprintf("%s", i) }
		logger = zerolog.New(out)
	}
	return logger.With().Str("unit", unit).Logger().Level(level)
}
