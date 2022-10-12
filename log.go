package jar

import (
	"github.com/cognusion/go-prw"
	"github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	// OutFormat is a log.Logger format used by default
	OutFormat = log.Ldate | log.Ltime | log.Lshortfile
	// DebugOut is a log.Logger for debug messages
	DebugOut = log.New(io.Discard, "[DEBUG] ", 0)
	// TimingOut is a log.Logger for timing-related debug messages. DEPRECATED
	TimingOut = log.New(io.Discard, "[TIMING] ", 0)
	// ErrorOut is a log.Logger for error messages
	ErrorOut = log.New(os.Stderr, "", OutFormat)
	// AccessOut is a log.Logger for access logging. PLEASE DO NOT USE THIS DIRECTLY
	AccessOut = log.New(os.Stdout, "", 0)
	// CommonOut is a log.Logger for Apache "common log format" logging. PLEASE DO NOT USE THIS DIRECTLY
	CommonOut = log.New(io.Discard, "", 0)
	// SlowOut is a log.Logger for slow request information
	SlowOut = log.New(io.Discard, "", 0)

	// RequestTimer is a function to allow Durations to be added to the Timer Metric
	RequestTimer func(time.Duration)

	// SlowRequests is the slow request log Duration
	SlowRequests time.Duration

	// LogPool is a Pool of AccessLogs
	LogPool sync.Pool
)

func init() {
	// init the Pool using JSONAccessLog. May be overridden later
	LogPool = sync.Pool{
		New: func() interface{} {
			return &JSONAccessLog{}
		},
	}

	// Register our request timer
	rt := metrics.NewRegisteredTimer("RequestTimes", Metrics)
	RequestTimer = func(d time.Duration) { rt.Update(d) }
}

// LogInit initializes all of the loggers based on Conf settings
func LogInit() error {

	// Set the Logs
	if !Conf.GetBool(ConfigCheckConfig) {
		ErrorOut = GetErrorLog(Conf.GetString(ConfigErrorLog), "", OutFormat, Conf.GetInt(ConfigLogSize), Conf.GetInt(ConfigLogBackups), Conf.GetInt(ConfigLogAge))
		AccessOut = GetLog(Conf.GetString(ConfigAccessLog), "", 0, Conf.GetInt(ConfigLogSize), Conf.GetInt(ConfigLogBackups), Conf.GetInt(ConfigLogAge))
		CommonOut = GetLogOrDiscard(Conf.GetString(ConfigCommonLog), "", 0, Conf.GetInt(ConfigLogSize), Conf.GetInt(ConfigLogBackups), Conf.GetInt(ConfigLogAge))
	}

	// Set the DebugOut, maybe
	if Conf.GetBool(ConfigDebug) {
		DebugOut = GetErrorLog(Conf.GetString(ConfigDebugLog), "[DEBUG] ", OutFormat, Conf.GetInt(ConfigLogSize), Conf.GetInt(ConfigLogBackups), Conf.GetInt(ConfigLogAge))
		Conf.Debug()
	}
	if Conf.GetBool(ConfigDebugTimings) {
		TimingOut = DebugOut
	}

	// Set the SlowRequest global, maybe
	if srm := Conf.GetString(ConfigSlowRequestMax); srm != "" {
		sr, err := time.ParseDuration(srm)
		if err != nil {
			return fmt.Errorf("error parsing slowrequestmax duration ('%s'): %w", srm, err)
		}
		SlowRequests = sr
		SlowOut = GetErrorLog(Conf.GetString(ConfigSlowLog), "", 0, Conf.GetInt(ConfigLogSize), Conf.GetInt(ConfigLogBackups), Conf.GetInt(ConfigLogAge))
	}

	return nil
}

// GetLog gets a standard-type log
func GetLog(filename, prefix string, format, size, backups, age int) *log.Logger {

	return getLog(filename, prefix, format, size, backups, age, os.Stdout)
}

// GetLogOrDiscard gets a standard-type log, or discards the output
func GetLogOrDiscard(filename, prefix string, format, size, backups, age int) *log.Logger {

	return getLog(filename, prefix, format, size, backups, age, io.Discard)
}

// GetErrorLog gets an error-type log
func GetErrorLog(filename, prefix string, format, size, backups, age int) *log.Logger {

	return getLog(filename, prefix, format, size, backups, age, os.Stderr)
}

// getLog abstracts all the things
func getLog(filename, prefix string, format, size, backups, age int, defaultWriter io.Writer) (l *log.Logger) {
	if filename == "" {
		// Nothing provided, use the defaults
		l = log.New(defaultWriter, prefix, format)
	} else {
		// File, use lumberjack
		l = log.New(&lumberjack.Logger{
			Filename:   filename,
			MaxSize:    size, // megabytes
			MaxBackups: backups,
			MaxAge:     age, // days
		}, prefix, format)
	}
	return
}

// loggerHook is a logrus.Hook so we can intercept oxy logging
type loggerHook struct {
	Name   string
	Log    *log.Logger
	levels []logrus.Level
}

func (l *loggerHook) AddLevel(level logrus.Level) {
	l.levels = append(l.levels, level)
}

func (l *loggerHook) AddLevels(levels []logrus.Level) {
	l.levels = append(l.levels, levels...)
}

func (l *loggerHook) Levels() []logrus.Level {
	return l.levels
}

func (l *loggerHook) Fire(entry *logrus.Entry) error {
	l.Log.Printf("%s: %s\n", l.Name, entry.Message)
	return nil
}

// AccessLog is an interface providing base logging, but allowing addons to extent it easily
type AccessLog interface {
	// CommonLogFormat will return the contents as a CLF-compatible string. If combined is set, a "combined" CLF is included (adds referer and user-agent)
	CommonLogFormat(combined bool) string
	// Reset will empty out the contents of the access log
	Reset()
	// ResponseFiller adds response information to the AccessLog entry
	ResponseFiller(responseTime time.Time, responseDuration time.Duration, responseCode int, responseLength int)
	// RequestFiller adds request information to the AccessLog entry
	RequestFiller(r *http.Request)
}

// JSONAccessLog is an AccessLog uberstruct for JSONifying log data
type JSONAccessLog struct {
	Timestamp     string `json:"timestamp"`
	Hostname      string `json:"hostname"`
	RemoteAddress string `json:"remoteaddr"`
	User          string `json:"user"`
	XForwardedFor string `json:"x-forwarded-for"`
	ClientIP      string `json:"clientip"`
	Method        string `json:"method"`
	Request       string `json:"request"`
	Status        string `json:"status"`
	Bytes         string `json:"bytes"`
	UserAgent     string `json:"user-agent"`
	Duration      string `json:"duration"`
	Referer       string `json:"referer"`
	Message       string `json:"message"`
	RequestID     string `json:"requestid"`
	Proto         string `json:"proto"`
	TLSVersion    string `json:"tlsversion"`
	clfTimestamp  string
}

// CommonLogFormat will return the contents as a CLF-compatible string. If combined is set, a "combined" CLF is included (adds referer and user-agent)
func (a *JSONAccessLog) CommonLogFormat(combined bool) string {
	var (
		address = "-"
		dashua  = "-"
		dashre  = "-"
		dashus  = "-"
	)
	if a.UserAgent != "" {
		dashua = a.UserAgent
	}
	if a.Referer != "" {
		dashre = a.Referer
	}
	if a.User != "" {
		dashus = a.User
	}
	if a.RemoteAddress != "" {
		address = ipOnly(a.RemoteAddress)
	}

	if combined {
		return fmt.Sprintf("%s - %s [%s] \"%s %s %s\" %s %s \"%s\" \"%s\"", address, dashus, a.clfTimestamp, a.Method, a.Request, a.Proto, a.Status, a.Bytes, dashre, dashua)
	}
	return fmt.Sprintf("%s - %s [%s] \"%s %s %s\" %s %s", address, dashus, a.clfTimestamp, a.Method, a.Request, a.Proto, a.Status, a.Bytes)
}

// Reset will empty out the contents of the access log
func (a *JSONAccessLog) Reset() {
	a.Timestamp = ""
	a.Hostname = ""
	a.RemoteAddress = ""
	a.User = ""
	a.XForwardedFor = ""
	a.ClientIP = ""
	a.Method = ""
	a.Request = ""
	a.Status = ""
	a.Bytes = ""
	a.UserAgent = ""
	a.Duration = ""
	a.Referer = ""
	a.Message = ""
	a.RequestID = ""
	a.clfTimestamp = ""
	a.Proto = ""
	a.TLSVersion = ""
}

// ResponseFiller adds response information to the AccessLog entry
func (a *JSONAccessLog) ResponseFiller(endtime time.Time, duration time.Duration, responseCode int, responseLength int) {
	a.Timestamp = endtime.Format(time.RFC3339Nano)
	a.clfTimestamp = endtime.Format("02/Jan/2006:15:04:05 -0700")

	a.Duration = strconv.FormatInt(duration.Nanoseconds(), 10)

	a.Status = strconv.Itoa(responseCode)
	a.Bytes = strconv.Itoa(responseLength)
}

// RequestFiller adds request information to the AccessLog entry
func (a *JSONAccessLog) RequestFiller(r *http.Request) {

	a.Proto = r.Proto
	a.Hostname = r.Host
	a.RemoteAddress = ipOnly(r.RemoteAddr)
	if u, _, ok := r.BasicAuth(); ok {
		a.User = u
	}
	a.Method = r.Method
	path := r.URL.Path
	if r.URL.RawQuery != "" {
		path += fmt.Sprintf("?%s", r.URL.RawQuery)
	}
	a.Request = path

	a.UserAgent = r.Header.Get("User-Agent")
	a.RequestID = r.Header.Get(Conf.GetString(ConfigRequestIDHeaderName))
	a.XForwardedFor = r.Header.Get("X-Forwarded-For")
	a.Referer = r.Referer()

	// If we're faking XFF for logging, and it's not already populated, do so
	if Conf.GetBool(ConfigLogFakeXFF) && a.XForwardedFor == "" {
		a.XForwardedFor = a.RemoteAddress
	}

	// TLS?
	if r.TLS != nil {
		a.TLSVersion = SslVersions.Suite(r.TLS.Version)
	}
}

// AccessLogHandler is a middleware that times how long requests takes, assembled an AccessLog, and logs accordingly
func AccessLogHandler(next http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {

		// Grab the current time
		starttime := time.Now().UTC()

		// Make a new PluggableResponseWriter if we need to
		DebugOut.Printf("AccessLogHandler Pluggable ResponseWriter...\n")
		rw, _ := prw.NewPluggableResponseWriterIfNot(w)
		defer rw.Flush()

		// Immediately pass on, and we'll handle the response headers at the end, tyvm
		next.ServeHTTP(rw, r)

		// status
		endtime := time.Now().UTC()
		//duration := time.Since(starttime)
		duration := endtime.Sub(starttime)

		// Call our RequestTimer metric
		RequestTimer(duration)

		// Grab some headers, maybe
		requestID := r.Header.Get(Conf.GetString(ConfigRequestIDHeaderName))

		// Get a logEntry
		logEntry := LogPool.Get().(AccessLog)
		defer LogPool.Put(logEntry)
		logEntry.Reset()
		logEntry.ResponseFiller(endtime, duration, rw.Code(), rw.Length())
		logEntry.RequestFiller(r)

		if Conf.GetBool(ConfigDebug) && Conf.GetBool(ConfigDebugResponses) {
			// dump the response, yo
			DebugOut.Printf("Response {%s} %+v\n", requestID, rw.Header())
		}

		j, _ := json.Marshal(logEntry)
		AccessOut.Println(string(j))
		CommonOut.Println(logEntry.CommonLogFormat(true))

		if SlowRequests != 0 && duration >= SlowRequests {
			SlowOut.Printf("{%s} took %s (%d)\n", requestID, duration.String(), duration.Nanoseconds())
		}
		TimingOut.Printf("{%s} Request took %s\n", requestID, duration.String())

	}
	return http.HandlerFunc(fn)
}
