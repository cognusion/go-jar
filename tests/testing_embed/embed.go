package main

import (
	"github.com/cognusion/go-jar"
	"github.com/spf13/pflag"

	"fmt"
	"net/http"
)

func init() {
	var err error

	// Set up the CLI for your program
	// pflag isn't required, but makes integration with
	// the Conf structure seamless (see Conf.BindPFlags, below)
	pflag.Bool(jar.ConfigDebug, false, "Enable vociferous output")
	pflag.Bool(jar.ConfigCheckConfig, false, "Run through the config load and then exit")
	config := pflag.String("config", "", "Config file to load")
	pflag.Parse()

	// Load the config file, maybe
	if *config != "" {
		err = jar.LoadConfig(*config, jar.Conf)
		if err != nil {
			jar.ErrorOut.Fatalf("Error loading config '%s': %s\n", *config, err)
		}
	}

	// Bind commandline flags to the global config
	jar.Conf.BindPFlags(pflag.CommandLine)

	jar.LogInit()
}

func main() {

	// Link our LazyLogger as a Handler
	jar.Handlers["lazylogger"] = LazyLogger

	// Link our AllDone as a Finisher
	jar.Finishers["donezo"] = AllDone

	// Make. It. So.
	jar.Bootstrap()

}

func AllDone(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Request Received")
}

func LazyLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("LAZY LOGGING: %+v\n", r)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
