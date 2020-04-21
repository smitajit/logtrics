-- script local variables can be defined here
local prefix = "logtrics.example.metrics"

-- logtrics instance to configure log parsing logic --
-- multiple logtrics instances can be configured in same script --
logtrics {
	-- optional --
	-- mainly used for logging purpose. But its better to name logtrics instances ---
	name = "example-logtrics",

	-- optional --
	-- to override default graphite configuration
	-- graphite =  {
		-- host = "127.0.0.1",
		-- port = 2003,
		-- interval = 2,
		-- debug = true
	-- },

	-- supports RE2 (https://en.wikipedia.org/wiki/RE2_(software)) regex for matching and substring extraction ---
	-- source, matched line and extracted substrings will be passed for process callback for metrics computation --
	expression = ".*",

	-- this callback function will be called for log line match based on the expression. ---
	-- @source : source of the log. e.g, console, upd:{host:port} ... ---
	-- @line : the log line ---
	-- @... : the substring matched by the regular expression ---
	process = function(source, line,  params, ...)
		local value = math.random(1,10)

		-- example logging apis --
		-- fatal("inside process. Source: %s , line: %s" , source , line)
		-- error("inside process. Source: %s , line: %s" , source , line)
		info("inside process. Source: %s , line: %s" , source , line)
		-- debug("inside process. Source: %s , line: %s" , source , line)
		-- trace("inside process. Source: %s , line: %s" , source , line)


		-- example graphite apis --
		-- graphite().counter(prefix .. ".counter.inc.value").inc(value)
		-- graphite().counter(prefix .. ".counter.dec.value").dec(value)
		-- graphite().timer(prefix .. ".timer.value").update(value)
		-- graphite().gauge(prefix .. ".gauge.value").update(value)
		-- graphite().meter(prefix .. ".meter.value").mark(value)
		end,
}

