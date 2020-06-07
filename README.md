# logtrics
application to read, aggregate, notify and generate metrics based on logs.

### configuration
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
  -m, --mode string             run modes, choices are "console", "udp", "tcp"'
  -d, --script.dir string       lua scripts directory (default "/etc/logtrics/scripts/")
  -f, --script.file string      lua script file path
      --tcp.host string         tcp server listening host (default "127.0.0.1")
      --tcp.port int            tcp server listening port (default 4003)
      --udp.host string         udp server listening host (default "127.0.0.1")
      --udp.port int            udp server listening port (default 4002)
  -v, --version                 version for logtrics
```
### modes

#### console
In this mode, the log lines can be provided to the console prompt to debug the scripts.
```
make
./logtrics -m console -f examples/demo.lua --logging.level debug
```
#### UDP
In this mode, the log lines can be read from the UDP socket
```
make
./logtrics -m udp -f examples/demo.lua --logging.level debug --udp.port 3002 --udp.host localhost
```
#### TCP
In this mode, the log lines can be read from the TCP socket
```
make
./logtrics -m tcp -f examples/demo.lua --logging.level debug --tcp.port 3002 --tcp.host localhost
```

### lua script
```lua
```
### RUN
* UDP
```
make
./logtrics -m udp -f examples/demo.lua --logging.level debug --udp.port 3002 --udp.host localhost
```

* Console
```bash
make
./logtrics -m console -f examples/demo.lua --logging.level debug
```
