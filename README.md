# logtrics
Applicaton to genrate mertrics from logs.

### configuration
```Usage:
  logtrics [flags]

Flags:
  -c, --config string              config file path (default "/etc/logtrics.toml")
      --graphite.host string       graphite server host
      --graphite.interval string   interval in secs
      --graphite.port string       graphite server port
  -h, --help                       help for logtrics
      --logging.level string       logging level (default "info")
      --logging.type string        logging type, choices are "syslog", "console" (default "console")
  -m, --mode string                run modes, choices are "console", "filetail", "udp", "tcp"'
  -d, --script.dir string          lua scripts directory
  -f, --script.file string         lua script file path
      --tcp.host string            tcp server listening host
      --tcp.port string            tcp server listening port
      --udp.host string            udp server listening host
      --udp.port string            udp server listening port
```
### lua script
```lua
-- script global variables can be defined here
-- local prefix = ""
logtrics {
	-- optional --
	-- can be used to override graphite configuration
	-- graphite =  {
		-- host = "127.0.0.1",
		-- port = 8080
	-- },
	expression = ".*",
	process = function(source , param1 , param2 , param3)
		local r = math.random(1,10)
		debug("process -> source: %s, param1: %s\n" , source , param1)
		-- graphite().counter(prefix).increment(r)
		end,
}
```
