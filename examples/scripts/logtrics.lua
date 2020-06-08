-- script local variables can be defined here
local prefix = "logtrics.example.metrics"

-- logtrics instance to configure log parsing logic --
-- multiple logtrics instances can be configured in same script --
logtrics {
	-- optional --
	-- mainly used for logging purpose. But its better to name logtrics instances ---
	name = "logtrics-example",

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
	-- expression for `hello "World"`. extracting word hello
	parser = {
		type = "re2",
		expression = 'hello "(?P<first>[a-zA-z0-9]+)"',
	},

	-- this callback function will be called for log line match based on the expression. ---
	-- @event : fields contains all the metainfo ---
	handler = function(event)
		local value = math.random(1,10)
		info("fields are %v" , event)
		-- if param1 == "world" then
			-- -- example logging apis --
			-- -- fatal("inside process. Source: %s , line: %s" , source , line)
			-- -- error("inside process. Source: %s , line: %s" , source , line)
			-- -- warn("inside process. Source: %s , line: %s" , source , line)
			-- info("found match. Match: %s, Source: %s , line: %s" , param1, source , line)
			-- -- debug("inside process. Source: %s , line: %s" , source , line)
			-- -- trace("inside process. Source: %s , line: %s" , source , line)
		-- else
			-- info("match not found. Source: %s , line: %s, param : %s" , source , line , param1)
		-- end


		-- example graphite apis --
		-- graphite().counter(prefix .. ".counter.inc.value").inc(value)
		-- graphite().counter(prefix .. ".counter.dec.value").dec(value)
		-- graphite().timer(prefix .. ".timer.value").update(value)
		-- graphite().gauge(prefix .. ".gauge.value").update(value)
		-- graphite().meter(prefix .. ".meter.value").mark(value)
		end,
}

