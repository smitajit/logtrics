// Package config is responsible for providing configuration
package config

import (
	"log/syslog"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog"
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
		Modes      []string  `toml:"modes"`
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
