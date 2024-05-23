package main

import (
	"github.com/cognusion/go-jar"
	"github.com/spf13/pflag"

	"fmt"
	"runtime/debug"
)

var (
	configVersion bool
	gopsAgent     bool

	tags string
)

func init() {
	var err error

	// Set up the CLI
	pflag.Bool(jar.ConfigDebug, false, "Enable vociferous output")
	pflag.Bool(jar.ConfigCheckConfig, false, "Run through the config load and then exit")
	pflag.Bool(jar.ConfigDumpConfig, false, "Load the config, dump it to stderr, and then exit")
	pflag.BoolVar(&configVersion, "version", false, "Print the version and then exit")
	pflag.BoolVar(&gopsAgent, "gopsagent", false, "Start the 'gops' agent")

	config := pflag.String("config", "", "Config file to load")
	pflag.Parse()

	// Load the config maybe
	err = jar.LoadConfig(*config, jar.Conf)
	if err != nil {
		jar.ErrorOut.Fatalf("Error loading config '%s': %s\n", *config, err)
	}

	// Bind commandline flags to viper config
	jar.Conf.BindPFlags(pflag.CommandLine)

	// Set up the the loggers
	if err := jar.LogInit(); err != nil {
		jar.ErrorOut.Fatalf("Error initializing logs: %s\n", err)
	}
}

func main() {
	if configVersion {
		// Pull the tags, if any, from the build
		bi, _ := debug.ReadBuildInfo()
		for _, v := range bi.Settings {
			if v.Key == "-tags" {
				tags = v.Value
				break
			}
		}

		fmt.Printf("JARD %s\nGo   %s\nCPUs %d\n",
			jar.VERSION,
			jar.GOVERSION,
			jar.NUMCPU)
		if tags != "" {
			// There are tags
			fmt.Printf("Tags %s\n", tags)
		}
		return
	}

	if jar.Conf.GetBool(jar.ConfigDumpConfig) {
		fmt.Println(jar.PrettyPrint(jar.Conf.AllSettings()))
		return
	}

	if gopsAgent {
		var psaChan = make(chan struct{})
		defer close(psaChan)
		startAgent(psaChan)
	}

	// Make. It. So.
	jar.Bootstrap()

}
