-- script global variables can be defined here
-- local prefix = ""

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
		debug("lua: processing log line [%s] from source [%s]" , line , source)
		-- graphite().counter(prefix).increment(r)
		end,
}

