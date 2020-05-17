package pkg

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/rs/zerolog"
	"github.com/smitajit/logtrics/pkg/config"
)

var (
	// ConsoleReaderPrompt is the prompt for console reader
	//nolint:gochecknoglobals
	ConsoleReaderPrompt = " logtrics \033[31m>\033[0m "

	//ConsoleReaderHistory is the history file for console input lines
	ConsoleReaderHistory = "/tmp/readline.tmp"

	// ConsoleReaderHelp is the help text to print on console reader startup
	//nolint:gochecknoglobals
	ConsoleReaderHelp = `
----------------------------------------------------------------------------------------
Console reader help text goes here
----------------------------------------------------------------------------------------
`
)

type (
	// ReadCallBackFun ...
	ReadCallBackFun = func(event LogEvent)

	// LogEvent is the event represents single log line read by the reader
	LogEvent struct {
		Source string
		Line   string
		Err    error
	}

	// LogReader is the interface to read logs
	LogReader interface {
		Start(ctx context.Context, cb ReadCallBackFun) error
	}

	// Console represents the LogReader in console mode
	Console struct {
		io.Writer
		io.Reader
		logger   zerolog.Logger
		readline *readline.Instance
	}

	// UDP represents the log reader in UDP server mode
	UDP struct {
		config *config.Configuration
		logger zerolog.Logger
	}
)

// NewConsole returns a new Console runner instance
func NewConsole(logger zerolog.Logger) (LogReader, error) {
	l, err := readline.NewEx(&readline.Config{
		Prompt:            ConsoleReaderPrompt,
		HistoryFile:       ConsoleReaderHistory,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		return nil, err
	}
	return &Console{Reader: os.Stdin, Writer: os.Stdout, logger: logger, readline: l}, nil
}

// Start the reader in console mode
func (c *Console) Start(ctx context.Context, cb ReadCallBackFun) error {
	fmt.Fprintln(c, ConsoleReaderHelp)
	for {
		select {
		case <-ctx.Done():
			c.logger.Debug().Msg("terminating console")
			return nil
		default:
			fmt.Print(ConsoleReaderPrompt)
			line, err := c.readline.Readline()
			if err == io.EOF {
				return nil
			}
			cb(LogEvent{"console", line, err})
		}
	}
}

// NewUDP returns a new reader which reads the logs from the UDP socket
func NewUDP(config *config.Configuration) LogReader {
	return &UDP{config: config, logger: config.Logger("reader: UDP")}
}

// Start starts the reader
func (s *UDP) Start(ctx context.Context, cb ReadCallBackFun) error {
	if s.config.UDP == nil || s.config.UDP.Host == "" {
		return fmt.Errorf("invalid UDP server configuration")
	}
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: s.config.UDP.Port,
		IP:   net.ParseIP(s.config.UDP.Host),
	})
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	s.logger.Debug().Msg("UDP server started")
	for {
		select {
		case <-ctx.Done():
			s.logger.Debug().Msg("UDP server terminated")
			return nil
		default:
			b := make([]byte, 1024)
			rlen, remote, err := conn.ReadFromUDP(b[:])
			if err != nil {
				cb(LogEvent{fmt.Sprintf("UDP:%s", remote), "", err})
			}
			line := strings.TrimSpace(string(b[:rlen]))
			line = strings.TrimSuffix(line, "\r\n")
			cb(LogEvent{fmt.Sprintf("UDP:%s", remote), line, nil})
		}
	}
}
