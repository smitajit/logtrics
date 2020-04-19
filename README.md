# logtrics
<span style="color:yellow">WIP</span> application to read, aggregate, notify and generate metrics based on logs.

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
-- script global variables can be defined here ---
-- useful to implement custom aggregrators ---
-- local prefix = "" ---

-- logtrics instance to configure log parsing logic --
-- multiple logtrics instances can be configured in same script --
logtrics {
	-- optional --
	-- can be used to override graphite configuration
	-- graphite =  {
		-- host = "127.0.0.1",
		-- port = 8080
	-- },

	-- supports RE2 (https://en.wikipedia.org/wiki/RE2_(software)) regex for matching and substring extraction ---
	-- source, matched line and extracted substrings will be passed for process callback for metrics computation --
	expression = ".*",

	-- this callback function will be called for log line match based on the expression. ---
	-- @source : source of the log. e.g, console, upd:{host:port} ... ---
	-- @line : the log line ---
	-- @... : the substring matched by the regular expression ---
	process = function(source, line,  params, ...)
		local r = math.random(1,10)

		--- example logging api to debug from script ---
		debug("lua: processing log line [%s] from source [%s]" , line , source)

		--- example graphite api to publish graphite metrics ---
		--- graphite().counter(prefix).increment(r) ---

		--- example notify api to send notifications ---
		--- notification().mail("recipient").subject("").body("").send() ---
		--- notification().slack("channel").message("").send() ---

		--- TODOs ---
		--- prometheus apis ---
		--- data persistence apis ---
		--- aggregation apis ---

		end,
}
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
