# logtrics

logtrics provide a way to parse logs, to generate metrics, notify and more.
It can read logs from multiple sources(console, UDP, TCP). It also provides interfaces through lua script to configure and customize your logging tricks :P

### Configuration

[sample](./examples/config.toml)

```Usage:
Usage:
  logtrics [flags]

Flags:
      --buffer.size int         go channel default buffer size
  -c, --config string           config file path (default "/etc/logtrics/logtrics.toml")
      --graphite.debug          if enabled metrics will be logged
      --graphite.host string    graphite server host (default "127.0.0.1")
      --graphite.interval int   interval in secs (default 30)
      --graphite.port int       graphite server port (default 2024)
  -h, --help                    help for logtrics
      --logging.level string    logging level (default "info")
      --logging.type string     logging type, choices are "syslog", "console" (default "console")
  -m, --modes strings           run modes, choices are "console", "udp", "tcp"'
  -d, --script.dir string       lua scripts directory (default "/etc/logtrics/scripts/")
  -f, --script.file string      lua script file path
      --tcp.host string         tcp server listening host (default "127.0.0.1")
      --tcp.port int            tcp server listening port (default 4003)
      --udp.host string         udp server listening host (default "127.0.0.1")
      --udp.port int            udp server listening port (default 4002)
  -v, --version                 version for logtrics
```

### Modes

    logtrics supports multiple mode to receive log line.

- console - Mainly for debugging scripts
- UDP/TCP - Receives logs using UDP/TCP socket. Mainly to be used with rsyslog [omfwd](https://www.rsyslog.com/doc/v8-stable/configuration/modules/omfwd.html)
- filetail - Receives logs by tailing log file. (TODO)

#### Console

In this mode, the log lines can be provided to the console prompt to debug the scripts.

```
logtrics -m console -f examples/scripts/logtrics.lua --logging.level debug
```

#### UDP

In this mode, the log lines can be read from the UDP socket

```
logtrics -m udp -f examples/scripts/logtrics.lua --logging.level debug --udp.port 4002 --udp.host localhost
```

send logs using `echo "hello \"World\"" | nc -cu localhost 4002`

#### TCP

In this mode, the log lines can be read from the TCP socket

```
logtrics -m tcp -f examples/scripts/logtrics.lua --logging.level debug --tcp.port 4003 --tcp.host localhost
```

send logs using `echo "hello \"World\"" | nc -c localhost 4003`

### Lua Script

[sample](./examples/scripts/logtrics.lua)

```lua
logtrics {
	name = "logtrics-example",
	parser = {
		type = "re2",
		-- expression for `hello "World"`. extracting word hello
		expression = 'hello "(?P<first>[a-zA-z0-9]+)"',
	},
	handler = function(fields)
		info("fields are %v" , fields)
		-- graphite().counter("demo.counter.inc.value").inc(value)
		-- graphite().counter("demo.counter.dec.value").dec(value)
		-- graphite().timer("demo.timer.value").update(value)
		-- graphite().gauge("demo.gauge.value").update(value)
		-- graphite().meter("demo.meter.value").mark(value)
		end,
}

```

### [TODO](./TODO.md)
