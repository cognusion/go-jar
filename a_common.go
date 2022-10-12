// Package jar is a readily-embeddable feature-rich proxy-focused AWS-aware
// distributed-oriented resiliency-enabling URL-driven superlative-laced
// ***elastic application link***. At its core, JAR is "just" a load-balancing
// proxy taking cues from HAProxy (resiliency, zero-drop restarts, performance)
// and Apache HTTPD (virtualize everything) while leveraging over 20 years
// of systems engineering experience to provide robust features with exceptional
// stability.
//
// JAR has been in production use since 2018 and handles millions of connections a day
// across heterogeneous application stacks.
//
// Consumers will want to 'cd cmd/jard; go build; #enjoy'
package jar

import (
	"github.com/cognusion/go-jar/aws"
	"github.com/cognusion/go-jar/funcregistry"
	"github.com/cognusion/go-jar/watcher"
	"github.com/cognusion/go-sequence"
	"github.com/cognusion/grace/gracehttp"
	"github.com/fsnotify/fsnotify"
	gerrors "github.com/go-errors/errors"
	"github.com/gorilla/mux"
	"github.com/mcuadros/go-version"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// ErrBootstrapDone should not be treated as a proper error, as it is returned if Bootstrap
	// is complete (e.g. checkconfig or doc output), and won't continue for non-error reasons
	ErrBootstrapDone = Error("Bootstrap() is done. This is not necessarily an error")

	// ErrPoolBuild is a panic in bootstrap() if BuildPools fails
	ErrPoolBuild = Error("error building pools")

	// ErrValidateExtras is a panic in bootstrap() if there are any errors in ValidateExtras.
	// Preceding error output may provide more specific information.
	ErrValidateExtras = Error("error validating extras")

	// ErrVersion is a panic in bootstrap() if versionrequired is set in the config, but is less
	// than the VERSION constant of the compiled binary reading the config.
	ErrVersion = Error("version requirement not met")
)

var (
	// Conf is the config struct
	Conf *viper.Viper

	// StopFuncs is an aggregator for functions that needs to be called during graceful shutdowns.
	// Can only be called once!
	StopFuncs = funcregistry.NewFuncRegistry(true)

	// StrainFuncs is an aggregator for functions that can be called when JAR is under resource pressure.
	StrainFuncs = funcregistry.NewFuncRegistry(false)

	// InitFuncs are called in the early phases of Bootstrap()
	InitFuncs = funcregistry.NewFuncRegistry(true)

	// FileWatcher is an abstracted mechanism for calling WatchHandlerFuncs when a file is changed
	FileWatcher *watcher.Watcher

	// LoadBalancers are Pools
	LoadBalancers *Pools

	// Seq is a Sequence used for request ids
	Seq = sequence.New(1)

	// Ec2Session is an aws.Session for use in various places
	Ec2Session *aws.Session

	// Hostname is a local cache of os.Hostname
	Hostname string
)

func init() {
	// Init the Config
	Conf = InitConfig()

	// Grab the hostname, maybe
	if me, err := os.Hostname(); err != nil {
		DebugOut.Printf("Error calling os.Hostname(): %s\n", err)
		Hostname = "localhost"
	} else {
		Hostname = me
	}

	// Setup the filewatcher
	{
		var err error
		FileWatcher, err = watcher.NewWatcher()
		if err != nil {
			ErrorOut.Fatalf("Error initializing watcher: %s\n", err)
		}
		StopFuncs.Add(func() {
			DebugOut.Printf("Stopping FileWatcher\n")
			// If we stop the FileWatcher, but keep running, we have a situation
			// that needs attention
			Status.Add("FileWatcher", "WARNING", "FileWatcher stopped", nil)
			FileWatcher.Close()
		})
	}

	InitFuncs.Add(func() {
		funcregistry.DebugOut = DebugOut

		watcher.ErrorOut = ErrorOut
		watcher.DebugOut = DebugOut
	})
}

func graceWrapper(servers []*http.Server, maxConnections int) chan error {
	errorChan := make(chan error, 1)

	go func() {
		var e error
		if maxConnections > 0 {
			e = gracehttp.ServeWithOptions(servers, gracehttp.ListenerLimit(maxConnections))
		} else {
			e = gracehttp.Serve(servers...)
		}
		errorChan <- e
	}()

	return errorChan
}

// BootstrapChan assumes that the Conf object is all set, for now at least,
// builds the necessary subsystems and starts running.
//
// BootstrapChan doesn't return unless the server exits or the passed chan is closed
func BootstrapChan(closer chan struct{}) {

	// bootstrap() will panic if it can't continue. We need to handle that.
	defer func() {
		if r := recover(); r != nil {
			ErrorOut.Println(r)
			panic(r)
		}
	}()

	done, servers := bootstrap()
	if done {
		return
	}

	// Runit
	gracehttp.SetLogger(ErrorOut)

	maxc := Conf.GetInt(ConfigMaxConnections)

	if maxc > 0 {
		DebugOut.Printf("Maximum connections set to %d\n", maxc)
	}

	errorChan := graceWrapper(servers, maxc)

	select {
	case e := <-errorChan:
		ErrorOut.Fatalf("%s exiting: %v\n", MacroDictionary.Replacer("%%NAME"), e)
	case <-closer:
		// We've been asked to exit
		KillSelf()
		return
	}

}

// ChanBootstrap assumes that the Conf object is all set, for now at least,
// builds the necessary subsystems and starts running.
//
// ChanBootstrap returns quickly, and should be assumed running unless an error
// is received on the returned chan. ErrBootstrapDone should not be treated as a
// proper error, as it is returned if Bootstrap is complete (e.g. checkconfig or doc output)
func ChanBootstrap() chan error {

	errorChan := make(chan error, 1)

	// bootstrap() will panic if it can't continue. We need to handle that.
	defer func() {
		if r := recover(); r != nil {
			var err error
			switch t := r.(type) {
			case string:
				err = Error(t)
			case error:
				err = t
			default:
				err = ErrUnknownError
			}
			errorChan <- err
		}
	}()

	done, servers := bootstrap()
	if done {
		errorChan <- ErrBootstrapDone
		return errorChan
	}

	go func() {
		// Runit
		gracehttp.SetLogger(ErrorOut)

		maxc := Conf.GetInt(ConfigMaxConnections)

		if maxc > 0 {
			DebugOut.Printf("Maximum connections set to %d\n", maxc)
		}

		graceErrorChan := graceWrapper(servers, maxc)

		e := <-graceErrorChan
		errorChan <- e
	}()

	return errorChan

}

// Bootstrap assumes that the Conf object is all set, for now at least,
// builds the necessary subsystems and starts running.
//
// Bootstrap doesn't return unless the server exits
func Bootstrap() {
	// bootstrap() will panic if it can't continue. We need to handle that.
	defer func() {
		if r := recover(); r != nil {
			if Conf.GetBool(ConfigRecovererLogStackTraces) {
				ErrorOut.Printf("Panic occurred: %s", gerrors.Wrap(r, 2).ErrorStack())
			}
			ErrorOut.Println(r)
			panic(r)
		}
	}()

	done, servers := bootstrap()
	if done {
		return
	}

	// Runit
	gracehttp.SetLogger(ErrorOut)

	if maxc := Conf.GetInt(ConfigMaxConnections); maxc > 0 {
		DebugOut.Printf("Maximum connections set to %d\n", maxc)
		ErrorOut.Fatalf("%s/%s exiting: %v\n", MacroDictionary.Replacer("%%NAME"), MacroDictionary.Replacer("%%VERSION"), gracehttp.ServeWithOptions(servers, gracehttp.ListenerLimit(maxc)))
	} else {
		// Default, no connection limit
		ErrorOut.Fatalf("%s/%s exiting: %v\n", MacroDictionary.Replacer("%%NAME"), MacroDictionary.Replacer("%%VERSION"), gracehttp.Serve(servers...))
	}
}

// bootstrap builds all the things, and returns a bool if it's complete and further execution is unneeded,
// and an array of http.Server that can be executed.
//
// Callers are advised that this function intentionally (or possibly unintentionally) will panic if it cannot
// continue the bootstrap process at any point. Callers are responsible for recovery.
func bootstrap() (done bool, servers []*http.Server) {

	// If we're going to use AWS/EC2 features, we need to turn this on early
	if Conf.GetBool(ConfigEC2) || Conf.GetString(ConfigKeysAwsAccessKey) != "" {
		aws.DebugOut = DebugOut
		aws.TimingOut = TimingOut

		var (
			awsRegion    = Conf.GetString(ConfigKeysAwsRegion)
			awsAccessKey = Conf.GetString(ConfigKeysAwsAccessKey)
			awsSecretKey = Conf.GetString(ConfigKeysAwsSecretKey)
			err          error
		)

		// TODO: config-driven params, vs. IAM instance profiles
		Ec2Session, err = aws.NewSession(awsRegion, awsAccessKey, awsSecretKey)
		if err != nil {
			panic(fmt.Errorf("error intializing AWS session: '%w'", err))
		}
	}

	// Run the InitFuncs
	InitFuncs.Call()

	// Let's get the version requirement out of the way first:
	if Conf.GetString(ConfigVersionRequired) != "" {
		normalVersion := version.Normalize(VERSION)
		normalRequired := version.Normalize(Conf.GetString(ConfigVersionRequired))
		if !version.Compare(normalVersion, normalRequired, ">=") {
			panic(fmt.Errorf("%w : %s (%s) < %s (%s)", ErrVersion, VERSION, normalVersion, Conf.GetString(ConfigVersionRequired), normalRequired))
		}
	}

	// Let's get Zulip out of the way
	if Conf.GetString(ConfigZulipBaseURL) != "" {
		ZulipClient = newZulipClient(Conf.GetString(ConfigZulipBaseURL), Conf.GetString(ConfigZulipUsername), Conf.GetString(ConfigZulipToken), Conf.GetInt(ConfigZulipRetryCount), Conf.GetDuration(ConfigZulipRetryInterval))
	}

	// r is our router. Long live r.
	r := mux.NewRouter()

	// Pools in jar, are groups of like-serving URIs.
	// Pools must be built before Paths, else (boom)
	{
		var ok bool
		if LoadBalancers, ok = BuildPools(); !ok {
			panic(ErrPoolBuild)
		}
	}

	// Paths in JAR, are like <location> maps
	if err := BuildPaths(r); err != nil {
		panic(fmt.Errorf("error creating paths: %w", err))
	}

	// If so configured, watch the config and restart if it changes
	if Conf.GetBool(ConfigHotConfig) {
		Conf.WatchConfig()
		Conf.OnConfigChange(func(e fsnotify.Event) {
			ErrorOut.Println("Config file changed:", e.Name)
			RestartSelf()
		})
	}

	// where the default listener listens
	listen := Conf.GetString(ConfigListen)

	// tls stuff
	var tlscfg *tls.Config
	if Conf.GetBool(ConfigTLSEnabled) {
		listen = Conf.GetString(ConfigTLSListen)
		var cl []uint16
		cl, err := Ciphers.CipherListToSuites(Conf.GetStringSlice(ConfigTLSCiphers))
		if err != nil {
			panic(fmt.Errorf("error creating TLS cipher suites: %w", err))
		}
		if len(cl) == 0 {
			// All of them.
			cl = Ciphers.AllSuites()
		}
		DebugOut.Printf("Ciphers:\n")
		for _, c := range Conf.GetStringSlice(ConfigTLSCiphers) {
			DebugOut.Printf("\t%s\n", c)
		}

		tlscfg = &tls.Config{
			PreferServerCipherSuites: true,
			CipherSuites:             cl,
		}

		// HTTP/2
		if Conf.GetBool(ConfigTLSHTTP2) {
			DebugOut.Println("HTTP2: true")
			tlscfg.NextProtos = append(tlscfg.NextProtos, "h2")
		}

		// Get the certs in there
		lc := Conf.GetStringMap(ConfigTLSCerts)
		certs := make([]Cert, len(lc))
		i := 0
		for k, c := range lc {
			cm := cast.ToStringMapString(c)
			certs[i] = Cert{
				Domain:   k,
				Keyfile:  cm["keyfile"],
				Certfile: cm["certfile"],
			}
			i++
		}

		// Load all the cert pairs
		DebugOut.Printf("Certs: %+v\n", certs)
		for _, c := range certs {
			cer, err := tls.LoadX509KeyPair(c.Certfile, c.Keyfile)
			if err != nil {
				panic(fmt.Errorf("error loading certpair %s %s: %w", c.Certfile, c.Keyfile, err))
			}
			tlscfg.Certificates = append(tlscfg.Certificates, cer)
		}

		var (
			minV = Conf.GetFloat64(ConfigTLSMinVersion)
			maxV = Conf.GetFloat64(ConfigTLSMaxVersion)
		)
		// Minimum TLS Verson switching. SSL not supported.
		switch {
		case minV < 1.1:
			tlscfg.MinVersion = tls.VersionTLS10

		case minV < 1.2:
			tlscfg.MinVersion = tls.VersionTLS11

		case minV < 1.3:
			tlscfg.MinVersion = tls.VersionTLS12

		default:
			tlscfg.MinVersion = tls.VersionTLS13
		}

		// Maximum TLS Verson switching. SSL not supported.
		switch {
		case maxV < 1.1:
			tlscfg.MaxVersion = tls.VersionTLS10

		case maxV < 1.2:
			tlscfg.MaxVersion = tls.VersionTLS11

		case maxV < 1.3:
			tlscfg.MaxVersion = tls.VersionTLS12

		default:
			tlscfg.MaxVersion = tls.VersionTLS13
		}

		DebugOut.Printf("TLSConfig preflight: Made it.\n")

	}

	if errs := ValidateExtras(); len(errs) > 0 {
		ErrorOut.Print("One or more errors validating configs:\n")
		for _, e := range errs {
			ErrorOut.Printf("\t%s\n", e)
		}
		panic(ErrValidateExtras)
	}

	// Compulsory redirects
	if Conf.GetBool(ConfigTLSEnabled) && Conf.GetBool(ConfigTLSHTTPRedirects) {
		DebugOut.Printf("Compulsory redirects...\n")

		// This Server is to redirect all requests made on the HTTP listener, to the HTTPS lisenter
		red := &http.Server{
			Addr:         Conf.GetString(ConfigListen),
			ReadTimeout:  time.Second * 30,
			WriteTimeout: time.Second * 30,
			IdleTimeout:  time.Millisecond * 5,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if autho := CheckAuthoritative(r); !autho {
					ErrorOut.Printf("Declining redirect of HTTP traffic from '%s' '%s' because not a valid authoritative domain\n", r.Host, r.URL.String())
					RequestErrorResponse(r, w, "Not authoritative for this domain", http.StatusBadRequest)
					return
				}

				newU := r.URL
				newU.Scheme = "https"
				h := r.Host
				hp := strings.Split(h, ":")
				hparts := strings.Split(listen, ":")
				newU.Host = fmt.Sprintf("%s:%s", hp[0], hparts[1])
				DebugOut.Printf("Redirecting HTTP traffic from '%s' to '%s'\n", Conf.GetString(ConfigListen), newU.String())
				http.Redirect(w, r, newU.String(), http.StatusMovedPermanently)
			}),
		}

		// append the redirect servers to the servers list
		servers = append(servers, red)
	}

	// Default listener
	s := &http.Server{
		Addr:        listen,
		Handler:     r,
		ErrorLog:    ErrorOut,
		TLSConfig:   tlscfg, // might be nil, might be live.
		IdleTimeout: Conf.GetDuration(ConfigKeepaliveTimeout),
		//ReadTimeout:  Conf.GetDuration(ConfigTimeout),
		//WriteTimeout: Conf.GetDuration(ConfigTimeout),
	}

	// We don't requre tls.enabled=true here, because it's a backdoor to disabling keepalives
	// even if TLS isn't being used, albeit clumsy.
	if Conf.GetBool(ConfigTLSKeepaliveDisabled) {
		s.SetKeepAlivesEnabled(false)
	}

	// append the main server to the servers list
	servers = append(servers, s)

	// Checkconfig bail before we spawn the listener
	if Conf.GetBool(ConfigCheckConfig) {
		DebugOut.Println("Checkconfig called, exiting...")
		fmt.Println("Ok")

		// We don't want to execute the servers, just validate the config, so we're done
		done = true
		return // os.Exit(0)
	}

	// We are ready to execute the servers
	return
}
