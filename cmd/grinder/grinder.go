package main

import (
	"github.com/cheggaaa/pb/v3"
	"github.com/cognusion/go-humanity"
	"github.com/cognusion/go-recyclable"
	"github.com/fatih/color"
	"github.com/viki-org/dnscache"
	"golang.org/x/net/context/ctxhttp"

	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
)

var (
	MaxRequests   int           // maximum number of outstanding HTTP get requests allowed
	Rounds        int           // How many times to hit it
	SleepTime     time.Duration // Duration to sleep between GETter spawns
	ErrOnly       bool          // Quiet unless 0 == Code >= 400
	NoColor       bool          // Disable colorizing
	NoDNSCache    bool          // Disable DNS caching
	Summary       bool          // Output final stats
	Save          bool          // Enable saving the file
	useBar        bool          // Use progress bar
	totalGuess    int           // Guesstimate of number of GETs (useful with -bar)
	debug         bool          // Enable debugging
	ResponseDebug bool          // Enable full response output if debug
	timeout       time.Duration // How long each GET request may take

	OutFormat = log.Ldate | log.Ltime | log.Lshortfile
	DebugOut  = log.New(io.Discard, "[DEBUG] ", OutFormat)

	BufferPool = recyclable.NewBufferPool()
)

type urlCode struct {
	URL  string
	Code int
	Size int64
	Dur  time.Duration
	Err  error
}

func init() {
	flag.IntVar(&MaxRequests, "max", 5, "Maximium in-flight GET requests at a time")
	flag.IntVar(&Rounds, "rounds", 100, "Number of times to hit the URL(s)")
	flag.BoolVar(&ErrOnly, "errorsonly", false, "Only output errors (HTTP Codes >= 400)")
	flag.BoolVar(&NoColor, "nocolor", false, "Don't colorize the output")
	flag.BoolVar(&Summary, "stats", false, "Output stats at the end")
	flag.DurationVar(&SleepTime, "sleep", 0, "Amount of time to sleep between spawning a GETter (e.g. 1ms, 10s)")
	flag.DurationVar(&timeout, "timeout", 0, "Amount of time to allow each GET request (e.g. 30s, 5m)")
	flag.BoolVar(&debug, "debug", false, "Enable debug output")
	flag.BoolVar(&ResponseDebug, "responsedebug", false, "Enable full response output if debugging is on")
	flag.BoolVar(&NoDNSCache, "nodnscache", false, "Disable DNS caching")
	flag.BoolVar(&useBar, "bar", false, "Use progress bar instead of printing lines, can still use -stats")
	flag.IntVar(&totalGuess, "guess", 0, "Rough guess of how many GETs will be coming for -bar to start at. It will adjust")
	flag.BoolVar(&Save, "save", false, "Save the content of the files. Into hostname/folders/file.ext files")
	flag.Parse()

	// Handle boring people
	if NoColor {
		color.NoColor = true
	}

	// Handle debug
	if debug {
		DebugOut = log.New(os.Stderr, "[DEBUG] ", OutFormat)
	}

	// Sets the default http client to use dnscache, because duh
	if !NoDNSCache {
		res := dnscache.New(1 * time.Hour)
		http.DefaultClient.Transport = &http.Transport{
			MaxIdleConnsPerHost: 64,
			Dial: func(network string, address string) (net.Conn, error) {
				separator := strings.LastIndex(address, ":")
				ip, _ := res.FetchOneString(address[:separator])
				return net.Dial("tcp", ip+address[separator:])
			},
		}
	}
}

func main() {

	var bar *pb.ProgressBar
	getChan := make(chan string, Rounds)        // Channel to stream URLs to get
	rChan := make(chan urlCode, MaxRequests*10) // Channel to stream responses from the Gets
	doneChan := make(chan bool)                 // Channel to signal a getter is done
	sigChan := make(chan os.Signal, 1)          // Channel to stream signals
	abortChan := make(chan bool)                // Channel to tell the getters to abort
	count := 0
	error4s := 0
	error5s := 0
	errors := 0
	mismatches := 0

	// Set up the progress bar
	if useBar {
		tmpl := `{{string . "prefix"}}{{counters . }} {{bar . }} {{percent . }} {{rtime . "ETA %s"}}{{string . "suffix"}}`
		bar = pb.ProgressBarTemplate(tmpl).New(totalGuess)
	}

	// Stream the signals we care about
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Signal handler
	go func() {
		<-sigChan
		DebugOut.Println("Signal seen, sending abort!")

		close(abortChan)
	}()

	// Spawn off the getters
	for g := 0; g < MaxRequests; g++ {
		go getter(getChan, rChan, doneChan, abortChan, timeout)
	}

	// Block until all the getters are done, and then close rChan
	go func() {
		defer close(rChan)

		for c := 0; c < MaxRequests; c++ {
			<-doneChan
			DebugOut.Printf("Done %d/%d\n", c+1, MaxRequests)
		}
	}()

	// spawn off the scanner
	start := time.Now()
	go scanStdIn(getChan, abortChan, bar)

	if useBar {
		bar.Start()
	}
	// Collate the results
	for i := range rChan {
		count++

		if useBar {
			bar.Increment()
		}
		if i.Code == 0 {
			errors++
			if useBar {
				continue
			}
			color.Red("%d (%s) %s %s (%s)\n", i.Code, humanity.ByteFormat(i.Size), i.URL, i.Dur.String(), i.Err)
		} else if i.Code < 400 {
			if ErrOnly || useBar {
				// skip
				continue
			}
			color.Green("%d (%s) %s %s\n", i.Code, humanity.ByteFormat(i.Size), i.URL, i.Dur.String())
		} else if i.Code < 500 {
			error4s++
			if useBar {
				continue
			}
			color.Yellow("%d (%s) %s %s\n", i.Code, humanity.ByteFormat(i.Size), i.URL, i.Dur.String())
		} else if i.Code < 600 {
			error5s++
			if useBar {
				continue
			}
			color.Red("%d (%s) %s %s\n", i.Code, humanity.ByteFormat(i.Size), i.URL, i.Dur.String())
		} else {
			mismatches++
			if useBar {
				continue
			}
			color.Red("%d (%s) %s %s\n", i.Code, humanity.ByteFormat(i.Size), i.URL, i.Dur.String())
		}
	}

	if useBar {
		bar.Finish()
	}
	elapsed := time.Since(start)

	if Summary {
		e := color.RedString("%d", errors)
		e4 := color.YellowString("%d", error4s)
		e5 := color.RedString("%d", error5s)
		eX := color.RedString("%d", mismatches)
		fmt.Printf("\n\nGETs: %d\nErrors: %s\n500 Errors: %s\n400 Errors: %s\nMismatches: %s\nElapsed Time: %s\n", count, e, e5, e4, eX, elapsed.String())
	}
}

// scanStdIn takes a channel to pass inputted strings to,
// and does so until EOF, whereafter it closes the channel
func scanStdIn(getChan chan string, abortChan chan bool, bar *pb.ProgressBar) {
	defer close(getChan)

	scanner := bufio.NewScanner(os.Stdin)
	count := int64(0)
	for scanner.Scan() {
		select {
		case <-abortChan:
			DebugOut.Println("scanner abort seen!")
			return
		default:
		}
		DebugOut.Println("scanner sending...")

		getURL := scanner.Text()

		for i := 0; i < Rounds; i++ {
			getChan <- getURL
			count++
			if bar != nil {
				if bar.Total() < count {
					bar.SetTotal(bar.Total() + 1)
				}
			}
		}

	}
	// POST: we've seen EOF
	DebugOut.Printf("EOF seen after %d lines\n", count)

}

func compare(resp *http.Response) bool {
	reqid := resp.Header.Get("X-Request-ID")
	reader := bufio.NewReader(resp.Body)
	line, _ := reader.ReadString(' ')
	line = strings.TrimSpace(line)
	DebugOut.Printf("%s == %s?\n", reqid, line)
	return line == reqid
}

// getter takes a receive channel, send channel, and done channel,
// running HTTP GETs for anything in the receive channel, returning
// formatted responses to the send channel, and signalling completion
// via the done channel
func getter(getChan chan string, rChan chan urlCode, doneChan chan bool, abortChan chan bool, timeout time.Duration) {
	defer func() { doneChan <- true }()

	var (
		ctx    context.Context
		cancel context.CancelFunc
		abort  bool
	)
	c := &http.Client{}

	go func() {
		// Wait until abort has been signalled,
		// then cancel all the things
		<-abortChan
		DebugOut.Println("getter abort seen!")
		abort = true
		if cancel != nil {
			cancel()
		}
	}()

	for url := range getChan {
		if abort {
			// Edge case: Abort has been called,
			// but we received a url via getChan
			DebugOut.Println("abort called")
			return
		} else if url == "" {
			// We assume an empty request is a closer
			// as that simplifies our for{select{}} loop
			// considerably
			DebugOut.Println("getter empty request seen!")
			return
		}
		DebugOut.Printf("getter getting %s\n", url)

		// Create the context
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(context.Background(), timeout)
		} else {
			ctx, cancel = context.WithCancel(context.Background())
		}

		// GET!
		s := time.Now()
		response, err := ctxhttp.Get(ctx, c, url)
		d := time.Since(s)

		if err != nil {
			// We assume code 0 to be a non-HTTP error
			rChan <- urlCode{url, 0, 0, d, err}
		} else {
			/*
				if ResponseDebug {
					b, err := io.ReadAll(response.Body)
					if err != nil {
						DebugOut.Printf("Error reading response body: %s\n", err)
					} else {
						DebugOut.Printf("<-----\n%s\n----->\n", b)
					}

					if Save {
						SaveFile(url, &b)
					}
				} else if Save {
					b, err := io.ReadAll(response.Body)
					if err != nil {
						fmt.Printf("Error reading response body: '%s' not saving file '%s'\n", err, url)
					} else {
						SaveFile(url, &b)
					}
				}
			*/
			cv := compare(response)
			if !cv {
				rChan <- urlCode{url, 611, response.ContentLength, d, nil}
			} else {
				rChan <- urlCode{url, response.StatusCode, response.ContentLength, d, nil}
			}
			response.Body.Close() // else leak
		}
		cancel()

		if abort {
			DebugOut.Println("abort called, post")
			return
		}

		if SleepTime > 0 {
			// Zzzzzzz
			time.Sleep(SleepTime)
		}
	}

}

// SaveFile takes a URL and a pointer to a []byte containing the to-be-saved bytes,
// and saves the full url as the path (sans scheme).
// e.g. 'https://somewhere.com/1/2/3/4/5.html' will be saved as './somewhere.com/1/2/3/4/5.html'
func SaveFile(saveAs string, contents *[]byte) error {
	url, err := url.Parse(saveAs)
	if err != nil {
		return err
	}

	dirs := path.Dir(url.Path)
	if !strings.HasPrefix(dirs, "/") {
		// Sanity!
		dirs = "/" + dirs
	}

	DebugOut.Printf("Saved File Path: '%s%s' full: '%s%s'\n", url.Hostname(), dirs, url.Hostname(), url.Path)
	err = os.MkdirAll(fmt.Sprintf("%s%s", url.Hostname(), dirs), os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s%s", url.Hostname(), url.Path), *contents, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
