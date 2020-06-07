// Package reader is responsible for reading log line from different data sources
package reader

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/rs/zerolog"
	"github.com/smitajit/logtrics/config"
)

var (
	// ConsoleReaderPrompt is the prompt for console reader
	//nolint:gochecknoglobals
	ConsoleReaderPrompt = " logtrics \033[31m>\033[0m "

	//ConsoleReaderHistory is the history file for console input lines
	//nolint:gochecknoglobals
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
	// ReadCallBack is the callback function for log events
	ReadCallBack = func(event LogEvent)

	// LogEvent is the event represents single log line read by the reader
	LogEvent struct {
		Source string
		Line   string
		Err    error
	}

	// LogReader is the interface to read logs
	LogReader interface {
		Start(ctx context.Context, cb ReadCallBack) error
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
		conf   *config.Configuration
		logger zerolog.Logger
	}

	// UDP represents the log reader in UDP server mode
	TCP struct {
		conf   *config.Configuration
		logger zerolog.Logger
	}
)

// NewConsole returns a new Console runner instance
func NewConsole(conf *config.Configuration) (LogReader, error) {
	l, err := readline.NewEx(&readline.Config{
		Prompt:            ConsoleReaderPrompt,
		EOFPrompt:         "exit",
		HistoryFile:       ConsoleReaderHistory,
		HistorySearchFold: true,
	})
	if err != nil {
		return nil, err
	}
	return &Console{Reader: os.Stdin, Writer: os.Stdout, logger: conf.Logger("reader: console"), readline: l}, nil
}

// Start the reader in console mode
func (c *Console) Start(ctx context.Context, cb ReadCallBack) error {
	fmt.Fprintln(c, ConsoleReaderHelp)
	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Debug().Msg("terminating console")
			default:
				fmt.Print(ConsoleReaderPrompt)
				line, err := c.readline.Readline()
				if err == io.EOF {
					return
				}
				fmt.Println(err)
				cb(LogEvent{"console", line, err})
			}
		}
	}()
	return nil
}

// NewUDP returns a new reader which reads the logs from the UDP socket
func NewUDP(conf *config.Configuration) LogReader {
	return &UDP{conf: conf, logger: conf.Logger("reader: UDP")}
}

// Start starts the reader
// this is a blocking call
func (s *UDP) Start(ctx context.Context, cb ReadCallBack) error {
	if s.conf.UDP == nil || s.conf.UDP.Host == "" {
		return fmt.Errorf("invalid UDP server configuration")
	}
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: s.conf.UDP.Port,
		IP:   net.ParseIP(s.conf.UDP.Host),
	})
	if err != nil {
		return err
	}
	s.logger.Debug().Msgf("UDP server started at [%s:%d]", s.conf.UDP.Host, s.conf.UDP.Port)
	go func() {
		defer func() { _ = conn.Close() }()
		for {
			select {
			case <-ctx.Done():
				s.logger.Debug().Msg("UDP server terminated")
				return
			default:
				b := make([]byte, 1024)
				len, remote, err := conn.ReadFromUDP(b)
				if err != nil {
					cb(LogEvent{fmt.Sprintf("UDP:%s", remote), "", err})
				}
				line := strings.TrimSpace(string(b[:len]))
				line = strings.TrimSuffix(line, "\r\n")
				cb(LogEvent{fmt.Sprintf("UDP:%s", remote), line, nil})
			}
		}
	}()
	return nil
}

// NewUDP returns a new reader which reads the logs from the UDP socket
func NewTCP(conf *config.Configuration) LogReader {
	return &TCP{conf: conf, logger: conf.Logger("reader: UDP")}
}

// Start starts the reader
// this is a blocking call
func (s *TCP) Start(ctx context.Context, cb ReadCallBack) error {
	if s.conf.TCP == nil || s.conf.TCP.Host == "" || s.conf.TCP.Port == 0 {
		return fmt.Errorf("invalid TCP server configuration")
	}

	addr := fmt.Sprintf("%s:%d", s.conf.TCP.Host, s.conf.TCP.Port)
	// Listen for incoming connections.
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// Close the listener when the application closes.
	s.logger.Debug().Msgf("TCP server started at [%s]", addr)
	go func() {
		defer l.Close()
		for {
			select {
			case <-ctx.Done():
				s.logger.Debug().Msg("terminating tcp server")
				return
			default:
				conn, err := l.Accept()
				if err != nil {
					s.logger.Error().Err(err).Msg("failed to accept tcp connection")
				}
				go func() {
					b := make([]byte, 1024)
					remote := conn.RemoteAddr().String()
					len, err := conn.Read(b)
					if err != nil {
						cb(LogEvent{fmt.Sprintf("TCP:%s", conn.RemoteAddr().String()), "", err})
					}
					line := strings.TrimSpace(string(b[:len]))
					line = strings.TrimSuffix(line, "\r\n")
					cb(LogEvent{fmt.Sprintf("TCP:%s", remote), line, nil})
				}()
			}
		}
	}()
	return nil
}
