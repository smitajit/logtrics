# logtrics
logtrics generates metrics by parsing regular expression.
It provides abstract APIs and lua binding to build parser and metrics generation logic

### Configuration
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

#### console
In this mode, the log lines can be provided to the console prompt to debug the scripts.
```
logtrics -m console -f examples/demo.lua --logging.level debug
```

#### UDP
In this mode, the log lines can be read from the UDP socket
```
logtrics -m udp -f examples/demo.lua --logging.level debug --udp.port 4002 --udp.host localhost
```
send logs using `echo "hello \"World\"" | nc -cu localhost 4002`

#### TCP
In this mode, the log lines can be read from the TCP socket
```
logtrics -m tcp -f examples/demo.lua --logging.level debug --tcp.port 4003 --tcp.host localhost
```
send logs using `echo "hello \"World\"" | nc -c localhost 4003`

### lua script
```lua
logtrics {
	name = "logtrics-example",
	parser = {
		type = "re2",
		-- expression for hello "World". extracting word hello
		expression = 'hello "(?P<first>[a-zA-z0-9]+)"', -- to parse hello "world"
	},
	handler = function(event)
		info("fields are %v" , event)
		-- graphite().counter(prefix .. ".counter.inc.value").inc(value)
		-- graphite().counter(prefix .. ".counter.dec.value").dec(value)
		-- graphite().timer(prefix .. ".timer.value").update(value)
		-- graphite().gauge(prefix .. ".gauge.value").update(value)
		-- graphite().meter(prefix .. ".meter.value").mark(value)
		end,
}

```
[sample](./examples/scripts/logtrics.lua)

### [TODO](./TODO.md)

