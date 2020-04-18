package reader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
)

var (
	// ConsoleReaderPrompt is prompt string
	ConsoleReaderPrompt = " logtrics > "

	// ConsoleReaderHelp is the help text:w
	ConsoleReaderHelp = `enter the log line by line
..............
Help text goes here
..............
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
	}
)

// NewConsole returns a new Console runner instance
func NewConsole() LogReader {
	return &Console{Reader: os.Stdin, Writer: os.Stdout}
}

// Start the reader in console mode
func (c *Console) Start(ctx context.Context, cb ReadCallBackFun) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprintln(c, ConsoleReaderHelp)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			fmt.Print(ConsoleReaderPrompt)
			line, err := reader.ReadString('\n')
			if err != nil {
				cb(LogEvent{"console", "", err})
				continue
			}
			cb(LogEvent{"console", line, nil})
		}
	}
}