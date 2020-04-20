package reader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/smitajit/logtrics/config"
)

var (
	// ConsoleReaderPrompt is prompt string
	//nolint:gochecknoglobals
	ConsoleReaderPrompt = " logtrics > "

	// ConsoleReaderHelp is the help text:w
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
		logger zerolog.Logger
	}

	// UDP represents the log reader in UDP server mode
	UDP struct {
		config *config.Configuration
		logger zerolog.Logger
	}
)

// NewConsole returns a new Console runner instance
func NewConsole(config *config.Configuration) LogReader {
	return &Console{Reader: os.Stdin, Writer: os.Stdout, logger: config.Logger("reader: console")}
}

// Start the reader in console mode
func (c *Console) Start(ctx context.Context, cb ReadCallBackFun) error {
	reader := bufio.NewReader(c.Reader)
	fmt.Fprintln(c, ConsoleReaderHelp)
	for {
		select {
		case <-ctx.Done():
			c.logger.Debug().Msg("terminating console")
			return nil
		default:
			fmt.Print(ConsoleReaderPrompt)
			line, err := reader.ReadString('\n')
			if err != nil {
				cb(LogEvent{"console", "", err})
				continue
			}
			line = strings.TrimRight(line, "\r\n")
			cb(LogEvent{"console", line, nil})
		}
	}
}

// NewUDP returns a new UDP server mode reader
func NewUDP(config *config.Configuration) LogReader {
	return &UDP{config: config, logger: config.Logger("reader: udp")}
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
			s.logger.Debug().Msg("terminating UDP server")
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
