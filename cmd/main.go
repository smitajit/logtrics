package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/smitajit/logtrics/config"
	"github.com/smitajit/logtrics/pkg"
	"github.com/smitajit/logtrics/pkg/reader"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// DefaultConfigPath is the default config path
	DefaultConfigPath = "/etc/logtrics.toml"
)

var (
	cmd = &cobra.Command{
		Use:   "logtrics",
		Short: "logtrics is a log parser metrics generator",
		Long: `logtrics generates metrics by parsing regular expression.
		it provides abstract APIs and lua binding to build parser and metrics generator logic`,
		RunE: func(_ *cobra.Command, _ []string) error { return run() },
	}
)

func init() {
	flags := cmd.PersistentFlags()

	flags.StringP("config", "c", DefaultConfigPath, "config file path")
	flags.StringP("mode", "m", "", `run modes, choices are "console", "filetail", "udp", "tcp"'`)
	flags.StringP("script.file", "f", "", "lua script file path")
	flags.StringP("script.dir", "d", "", "lua script dir path")
	flags.StringP("logging.level", "", "info", "logging level")
	flags.StringP("logging.type", "", "console", `logging type, choices are "syslog", "console"`)

	_ = viper.BindPFlag("config", flags.Lookup("config"))
	_ = viper.BindPFlag("mode", flags.Lookup("mode"))
	_ = viper.BindPFlag("scriptfile", flags.Lookup("script.file"))
	_ = viper.BindPFlag("scriptdir", flags.Lookup("script.dir"))
	_ = viper.BindPFlag("logging.level", flags.Lookup("logging.level"))
	_ = viper.BindPFlag("logging.type", flags.Lookup("logging.type"))

	cobra.OnInitialize(func() {
		viper.SetConfigFile(viper.GetString("config"))
		if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
			panic(err)
		}
	})
}

func run() error {
	config := &config.Configuration{}
	if err := viper.Unmarshal(config); err != nil {
		return err
	}
	fmt.Printf("the config is %+v\n", config)
	switch config.Mode {
	case "console":
		return runConsole(config)
	case "filetail":
	case "udp":
	case "tcp":
	default:
		return fmt.Errorf(`invalid application mode. Choices are "console", "filetail", "tcp", "udp" `)
	}
	return nil
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
	fmt.Printf("script file are %+v \n", scripts) // TODO log
	return scripts, err
}

func runConsole(config *config.Configuration) error {
	scripts, err := scripts(config)
	if nil != err {
		return errors.Wrap(err, "failed to get the scripts")
	}
	if len(scripts) == 0 {
		return fmt.Errorf("no scripts found")
	}
	reader := reader.NewConsole()
	app, err := pkg.NewApp(config, reader, scripts...)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Start(true); err != nil {
		log.Fatal(err)
	}
	return nil
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
