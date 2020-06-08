package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/smitajit/logtrics"
	"github.com/smitajit/logtrics/config"
	"github.com/smitajit/logtrics/reader"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// defaultConfigPath is the default config path
	defaultConfigPath = "/etc/logtrics/config.toml"

	//defaultScriptDir is the default location from where all the scripts will be read
	defaultScriptDir = "/etc/logtrics/scripts/"
)

var (
	//nolint:gochecknoglobals
	cmd = &cobra.Command{
		Use:     "logtrics",
		Short:   "logtrics provide a way to parse logs, to generate metrics, notify and more",
		Version: logtrics.Version,
		Long:    "logtrics provide a way to parse logs, to generate metrics, notify and more",
		RunE:    func(_ *cobra.Command, _ []string) error { return run() },
	}
)

//nolint:gochecknoinits
func init() {
	flags := cmd.PersistentFlags()

	flags.StringP("config", "c", defaultConfigPath, "config file path")
	flags.StringSliceP("modes", "m", []string{}, `comma separated run modes, choices are "console", "udp", "tcp"'`)
	flags.Int("buffer.size", 0, "go channel default buffer size")

	flags.StringP("script.file", "f", "", "lua script file path")
	flags.StringP("script.dir", "d", defaultScriptDir, "lua scripts directory")

	flags.String("logging.level", "info", "logging level")
	flags.String("logging.type", "console", `logging type, choices are "syslog", "console"`)

	flags.String("udp.host", "127.0.0.1", "udp server listening host")
	flags.Int("udp.port", 4002, "udp server listening port")

	flags.String("tcp.host", "127.0.0.1", "tcp server listening host")
	flags.Int("tcp.port", 4003, "tcp server listening port")

	flags.String("graphite.host", "127.0.0.1", "graphite server host")
	flags.Int("graphite.port", 2024, "graphite server port")
	flags.Int("graphite.interval", 30, "interval in secs")
	flags.Bool("graphite.debug", false, "if enabled metrics will be logged")

	_ = viper.BindPFlag("config", flags.Lookup("config"))
	_ = viper.BindPFlag("modes", flags.Lookup("modes"))
	_ = viper.BindPFlag("buffersize", flags.Lookup("buffer.size"))
	_ = viper.BindPFlag("scriptfile", flags.Lookup("script.file"))
	_ = viper.BindPFlag("scriptdir", flags.Lookup("script.dir"))
	_ = viper.BindPFlag("logging.level", flags.Lookup("logging.level"))
	_ = viper.BindPFlag("logging.type", flags.Lookup("logging.type"))
	_ = viper.BindPFlag("udp.port", flags.Lookup("udp.port"))
	_ = viper.BindPFlag("udp.host", flags.Lookup("udp.host"))
	_ = viper.BindPFlag("tcp.port", flags.Lookup("tcp.port"))
	_ = viper.BindPFlag("tcp.host", flags.Lookup("tcp.host"))
	_ = viper.BindPFlag("graphite.host", flags.Lookup("graphite.host"))
	_ = viper.BindPFlag("graphite.port", flags.Lookup("graphite.port"))
	_ = viper.BindPFlag("graphite.interval", flags.Lookup("graphite.interval"))
	_ = viper.BindPFlag("graphite.debug", flags.Lookup("graphite.debug"))

	cobra.OnInitialize(func() {
		viper.SetConfigFile(viper.GetString("config"))
		if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
			panic(err)
		}
	})
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config := &config.Configuration{}
	if err := viper.Unmarshal(config); err != nil {
		return err
	}
	if len(config.Modes) == 0 {
		return errors.New("need atleast one application mode")
	}

	var readers []reader.LogReader
	for _, m := range config.Modes {
		switch m {
		case "console":
			reader, err := reader.NewConsole(config)
			if err != nil {
				return err
			}
			readers = append(readers, reader)
		case "udp":
			reader := reader.NewUDP(config)
			readers = append(readers, reader)
		case "tcp":
			reader := reader.NewTCP(config)
			readers = append(readers, reader)
		default:
			return fmt.Errorf(`invalid application mode. Choices are "console", "tcp", "udp" `)
		}
	}

	app, err := logtrics.NewApplication(config, readers...)
	if err != nil {
		return err
	}
	if err := app.Run(ctx); err != nil {
		return err
	}
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	return nil
}

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
