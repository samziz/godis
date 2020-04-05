math.randomseed(os.time())

request = function()
   wrk.method = "GET"
   wrk.body = string.format('{"op": "SET", "key": "%s", "value": "%s"}', math.random(10000), math.random(10000))
   return wrk.format("GET", "/")
end