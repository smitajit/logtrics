package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/smitajit/logtrics/config"
	"github.com/smitajit/logtrics/pkg"
	"github.com/smitajit/logtrics/pkg/reader"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// DefaultConfigPath is the default config path
	DefaultConfigPath = "/etc/logtrics/logtrics.toml"

	//DefaultScriptDir is the default location from where all the scripts will be read
	DefaultScriptDir = "/etc/logtrics/scripts/"
)

var (
	//nolint:gochecknoglobals
	cmd = &cobra.Command{
		Use:     "logtrics",
		Short:   "logtrics is a log parser metrics generator",
		Version: "1.0.0",
		Long: `logtrics generates metrics by parsing regular expression.
		it provides abstract APIs and lua binding to build parser and metrics generator logic`,
		RunE: func(_ *cobra.Command, _ []string) error { return run() },
	}
)

func init() {
	flags := cmd.PersistentFlags()

	flags.StringP("config", "c", DefaultConfigPath, "config file path")
	flags.StringP("mode", "m", "", `run modes, choices are "console", "filetail", "udp", "tcp"'`)
	flags.Int("buffer.size", 0, "go channel default buffer size")

	flags.StringP("script.file", "f", "", "lua script file path")
	flags.StringP("script.dir", "d", DefaultScriptDir, "lua scripts directory")

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
	_ = viper.BindPFlag("mode", flags.Lookup("mode"))
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
	// adding interrupt handler
	go func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()

	config := &config.Configuration{}
	if err := viper.Unmarshal(config); err != nil {
		return err
	}
	switch config.Mode {
	case "console":
		return runConsole(ctx, config)
	case "filetail":
		return fmt.Errorf("not implemented yet")
	case "udp":
		return runUDP(ctx, config)
	case "tcp":
		return fmt.Errorf("not implemented yet")
	default:
		return fmt.Errorf(`invalid application mode. Choices are "console", "filetail", "tcp", "udp" `)
	}
}

func scripts(config *config.Configuration) ([]string, error) {
	if config.ScriptFile != "" {
		return []string{config.ScriptFile}, nil
	}
	scripts := make([]string, 0)
	err := filepath.Walk(config.ScriptDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".lua") {
			scripts = append(scripts, path)
		}
		return nil
	})
	if len(scripts) == 0 {
		return nil, fmt.Errorf("no scripts found")
	}
	return scripts, err
}

func runConsole(ctx context.Context, conf *config.Configuration) error {
	scripts, err := scripts(conf)
	if err != nil {
		return err
	}
	reader, err := reader.NewConsole(conf)
	if err != nil {
		return err
	}
	app, err := pkg.NewApplication(conf, reader, scripts...)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
	return nil
}

func runUDP(ctx context.Context, conf *config.Configuration) error {
	scripts, err := scripts(conf)
	if nil != err {
		return err
	}
	reader := reader.NewUDP(conf)
	app, err := pkg.NewApplication(conf, reader, scripts...)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.RunAsync(ctx); err != nil {
		log.Fatal(err)
	}
	return nil
}

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
