package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/cognusion/srvdisco"
	"github.com/spf13/pflag"
)

var (
	debug      bool
	useSRV     bool
	skipVerify bool
	suffixSRV  string
	domainSRV  string
	scheme     string
	targets    string
	baseURI    string
	command    string
	pool       string

	commands = []string{"add", "lose", "list"}
)

const (
	uaversion = "1.2"
	prefixSRV = "jarpool"
)

func init() {
	// Set up the CLI
	pflag.BoolVar(&debug, "debug", false, "Enable vociferous output")
	pflag.BoolVar(&useSRV, "srv", true, "Use DNS SRV to look up targets")
	pflag.BoolVar(&skipVerify, "skipverify", true, "Skip certificate verification")
	pflag.StringVar(&suffixSRV, "srvsuffix", "", "Suffix of DNS SRV to use (e.g. dev, prod, useast1c, etc.)")
	pflag.StringVar(&domainSRV, "srvdomain", "", "Domain of DNS SRV to use")
	pflag.StringVar(&scheme, "scheme", "https", "Protocol scheme to prefix")
	pflag.StringVar(&baseURI, "uri", "/admin/pool", "Base path to the manager")
	pflag.StringVar(&targets, "targets", "", "List of JARDs to manage (if not using SRV)")
	pflag.StringVar(&pool, "pool", "", "Pool to act upon")
	pflag.StringVar(&command, "command", "", fmt.Sprintf("Command to issue, one of: %v", commands))
	pflag.Parse()

	if skipVerify {
		// DGAF about cert issues. Make It So.
		//#nosec G402 -- Tool explicit skipverify.
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
}

func get(url string) error {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("PoolManager/%s", uaversion))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n%s\n", url, string(body))
	return nil

}

func main() {

	var jards []string

	if useSRV {
		var err error
		jards, err = srvdisco.Discover(domainSRV, fmt.Sprintf("%s%s", prefixSRV, suffixSRV), scheme)
		if err != nil {
			fmt.Printf("Error during SRV lookup: %s\n", err.Error())
			os.Exit(1)
		}

	} else {
		// Not using SRV: Hardcode

		if targets == "" {
			fmt.Println("Either -srv must be set, or -targets must not be empty")
			os.Exit(1)
		}

		jards = strings.Split(targets, ",")
		for i, s := range jards {
			jards[i] = fmt.Sprintf("%s://%s", scheme, strings.TrimSpace(s))
		}
	}

	fmt.Println("Targets:")
	for _, jard := range jards {
		fmt.Printf("\t%s%s\n", jard, baseURI)
	}
	fmt.Println()

	// Handle commands
	commandOk := false
	commandFields := strings.Fields(command)

	if len(commandFields) == 0 {
		fmt.Println("Must set --command")
		os.Exit(1)
	}

	for _, c := range commands {
		if c == commandFields[0] {
			commandOk = true
			break
		}
	}
	if !commandOk {
		// Wasn't in list
		fmt.Printf("Invalid --command '%s'\n", commandFields[0])
		os.Exit(1)
	} else if commandFields[0] != "list" && len(commandFields) != 2 {
		// Needs unprovided argument
		fmt.Printf("Command '%s' requires an additional argument\n", commandFields[0])
		os.Exit(1)
	}

	if commandFields[0] != "list" && pool == "" {
		fmt.Println("Must set --pool unless --command list")
		os.Exit(1)
	}

	var uri string
	if commandFields[0] == "list" && pool == "" {
		uri = fmt.Sprintf("%s/%s", baseURI, commandFields[0])
	} else if commandFields[0] == "list" {
		uri = fmt.Sprintf("%s/%s/%s", baseURI, pool, commandFields[0])
	} else {
		// The second command argument needs base64-encoding
		b64thing := base64.StdEncoding.EncodeToString([]byte(commandFields[1]))
		uri = fmt.Sprintf("%s/%s/%s/%s", baseURI, pool, commandFields[0], b64thing)
	}

	for _, jard := range jards {
		url := fmt.Sprintf("%s%s", jard, uri)
		err := get(url)
		if err != nil {
			fmt.Printf("Error calling GET %s : %s\n", url, err.Error())
		}
	}
	fmt.Println("Complete")
}
