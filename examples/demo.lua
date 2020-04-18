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
