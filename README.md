An experimental Redis clone in Go just for explaining some basic Go stuff

### To run

Start a server on port 9010 (why not? that's a respectable-sounding port):

```bash
$ PORT=9010 go run ./src
```

And in another terminal, make a couple of calls and pipe the JSON
response into JQ to prettify it:

```bash
$ curl -d '{"op": "SET", "key": "foo", "value": "bar"}' 127.0.0.1:9010 2>/dev/null | jq
{
  "Status": 200
}
$ curl -d '{"op": "GET", "key": "foo"}' 127.0.0.1:9010 2>/dev/null | jq                
{
  "Status": 200,
  "Value": "bar"
}
```

If we run the load-testing script, we see we can handle 
a throughput of about 40k requests / sec on a MacBook Pro,
which is decent:

```shell script
$ ./loadtesting/run.sh       
Running 10s test @ http://localhost:9010
  20 threads and 5000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    11.64ms   18.49ms   1.34s    99.94%
    Req/Sec     2.41k   789.68     8.59k    85.90%
  Latency Distribution
     50%   11.30ms
     75%   17.12ms
     90%   20.15ms
     99%   24.62ms
  394111 requests in 10.10s, 49.24MB read
  Socket errors: connect 0, read 1377, write 0, timeout 0
Requests/sec:  39004.51
Transfer/sec:      4.87MB
```