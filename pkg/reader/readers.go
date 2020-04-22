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
	ConsoleReaderPrompt = " logtrics \033[31m»\033[0m "

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
		logger zerolog.Logger
		l      *readline.Instance
	}

	// UDP represents the log reader in UDP server mode
	UDP struct {
		config *config.Configuration
		logger zerolog.Logger
	}
)

// NewConsole returns a new Console runner instance
func NewConsole(config *config.Configuration) (LogReader, error) {
	l, err := readline.NewEx(&readline.Config{
		Prompt:            ConsoleReaderPrompt,
		HistoryFile:       "/tmp/readline.tmp",
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		return nil, err
	}
	return &Console{Reader: os.Stdin, Writer: os.Stdout, logger: config.Logger("reader: console"), l: l}, nil
}

// Start the reader in console mode
func (c *Console) Start(ctx context.Context, cb ReadCallBackFun) error {
	// reader := bufio.NewReader(c.Reader)
	fmt.Fprintln(c, ConsoleReaderHelp)
	for {
		select {
		case <-ctx.Done():
			c.logger.Debug().Msg("terminating console")
			return nil
		default:
			fmt.Print(ConsoleReaderPrompt)
			// line, err := reader.ReadString('\n')
			// if err != nil {
			// cb(LogEvent{"console", "", err})
			// continue
			// }
			// line = strings.TrimRight(line, "\r\n")
			line, err := c.l.Readline()
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
