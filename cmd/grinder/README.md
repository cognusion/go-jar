# grinder

Grinder is a trivial load-tester.

e.g.

`echo http://localhost:8080/ | ./grinder -bar -stats -max 2000 -rounds 100000`

"Hit http://localhost:8080/, 100,000 times, at most 2,000 parallel requests, use a progress bar, and show me stats at the end"

```
Usage of ./grinder:
  -bar
    	Use progress bar instead of printing lines, can still use -stats
  -debug
    	Enable debug output
  -errorsonly
    	Only output errors (HTTP Codes >= 400)
  -guess int
    	Rough guess of how many GETs will be coming for -bar to start at. It will adjust
  -max int
    	Maximium in-flight GET requests at a time (default 5)
  -nocolor
    	Don't colorize the output
  -nodnscache
    	Disable DNS caching
  -responsedebug
    	Enable full response output if debugging is on
  -rounds int
    	Number of times to hit the URL(s) (default 100)
  -sleep duration
    	Amount of time to sleep between spawning a GETter (e.g. 1ms, 10s)
  -stats
    	Output stats at the end
  -timeout duration
    	Amount of time to allow each GET request (e.g. 30s, 5m)
```