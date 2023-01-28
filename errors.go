package jar

import (
	"github.com/cognusion/go-prw"
	"github.com/cognusion/go-timings"

	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	errorWrappedKey errorsID = iota

	// ErrPreviousError is exclusive to the HandleSuddenEviction scheme,
	// and announces that original error should be the one returned to the caller
	ErrPreviousError = Error("previous error stands")

	// ErrAuthError is used to communicate authentication errors. More detail
	// will be in the error log, but let's not leak that, shall we?
	ErrAuthError = Error("an error occurred during AAA")

	// ErrForbiddenError is used to communicate resource access denial
	ErrForbiddenError = Error("you do not have access to this resource")

	// ErrNoSession is called when an AWS feature is called, but there is no initialized AWS session
	ErrNoSession = Error("there is no initialized AWS session")

	// ErrUnknownError is returned when an error occurs for undefined-yet-anticipated reasons
	ErrUnknownError = Error("unknown error")
)

// Constants for configuration key strings
const (
	ConfigErrorHandlerTemplate = ConfigKey("errorhandler.template")
	ConfigErrorHandlerURI      = ConfigKey("errorhandler.uri")
)

// Constants for miscellaneous conditions
const (
	StatusClientClosedRequest     = 499
	StatusClientClosedRequestText = "Client Closed Request"
)

// Error is an error type
type Error string

// Error returns the stringified version of Error
func (e Error) Error() string {
	return string(e)
}

var (
	// ErrorTemplate is an HTML template for returning errors
	ErrorTemplate *template.Template

	errorWrapper ErrorWrapper
)

func init() {
	InitFuncs.Add(func() {
		if templ := Conf.GetString(ConfigErrorHandlerTemplate); templ != "" {
			var err error
			ErrorTemplate, err = template.ParseFiles(templ)
			if err != nil {
				ErrorOut.Fatalf("Unable to parse error template '%s': %s", templ, err)
			}
			errorWrapper = ErrorWrapper{HandleTemplateWrapper}
			Handlers["errorhandler"] = errorWrapper.Handler
		} else if uri := Conf.GetString(ConfigErrorHandlerURI); uri != "" {
			errorWrapper = ErrorWrapper{HandleRemoteWrapper}
			Handlers["errorhandler"] = errorWrapper.Handler
		} else {
			errorWrapper = ErrorWrapper{HandleGenericWrapper}
			Handlers["errorhandler"] = errorWrapper.Handler
		}
	})
}

type errorsID int

// ErrConfigurationError is returned when a debilitating configuration error
// occurs. If this is the initial configuration load, the program should exit.
// If this is a reload, the reload should abort and the known-working configuration
// should persist
type ErrConfigurationError struct {
	message string
}

func (e ErrConfigurationError) Error() string {
	if e.message != "" {
		return e.message
	}
	return "a configuration error occurred"
}

// ErrRequestError should be returned whenever an error is returned to
// a requestor. Care should be taken not to expose dynamic information inside
// the message. The request id will be automatically added to the message
type ErrRequestError struct {
	Request *http.Request
	Message string
}

// RequestErrorString is the functional equivalent of ErrRequestError .String()
func RequestErrorString(Request *http.Request, Message string) string {
	return ErrRequestError{Request, Message}.String()
}

// RequestErrorResponse is the functional equivalent of ErrRequestError .WrappedResponse(..)
func RequestErrorResponse(r *http.Request, w http.ResponseWriter, Message string, code int) {
	ErrRequestError{r, Message}.WrappedResponse(code, w)
}

// Bytes returns a []byte of the error
func (e ErrRequestError) Bytes() []byte {
	return []byte(e.Error())
}

// String returns a string of the error
func (e ErrRequestError) String() string {
	return e.Error()
}

// Error returns a string of the error
func (e ErrRequestError) Error() string {
	var requestID string
	if rid := e.Request.Context().Value(requestIDKey); rid != nil {
		requestID = rid.(string)
	}

	if e.Message != "" {
		return fmt.Sprintf("{%s} %s", requestID, e.Message)
	}
	return fmt.Sprintf("{%s} %s", requestID, "a configuration error occurred")
}

// WrappedResponse writes the templatized version of the error to a PRW
func (e ErrRequestError) WrappedResponse(code int, w http.ResponseWriter) {
	rw, _ := prw.NewPluggableResponseWriterIfNot(w)
	defer rw.Flush()

	errorWrapper.E(code, e.Request, rw, e.Bytes())
}

// TemplateError is a static structure to pass into error-wrapping templates
type TemplateError struct {
	// ErrorCode is the string value of the error
	ErrorCode string
	// ErrorMessage is an optional message the template may optionally render
	ErrorMessage string
	// RedirectURL is a URL the template is advised to redirect to
	RedirectURL string
	// RedirectSeconds is the number of seconds the template is advised to wait
	// before executing the RedirectURL
	RedirectSeconds int
}

// An ErrorWrapper is a struct to abstract error wrapping
type ErrorWrapper struct {
	// E takes the error code, request, a PluggableResponseWriter, and the original body,
	// and returns boolean true IFF rw has been written to. E should not change
	// headers as they may be ignored.
	E func(code int, r *http.Request, rw *prw.PluggableResponseWriter, body []byte) bool
}

// Handler is the chainable handler that will wrap the error
func (e *ErrorWrapper) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		// Make a new PluggableResponseWriter if we need to
		DebugOut.Printf("ErrorWrapper.Handler Pluggable ResponseWriter...\n")
		rw, _ := prw.NewPluggableResponseWriterIfNot(w)
		defer rw.Flush()

		// Immediately pass on, and we'll handle the response headers at the end, tyvm
		next.ServeHTTP(rw, r)

		if rw.Code() >= 400 {
			// An error we might care about.

			// Timings
			t := timings.Tracker{}
			t.Start()

			var errorWrapped bool
			if ew := r.Context().Value(errorWrappedKey); ew != nil {
				// if we've already wrapped the error, we don't need to do it again
				errorWrapped = true
			}
			if r.URL.Path == "/messages.php" {
				// chances are, this is a wrapped error page
				errorWrapped = true
			}

			// if we haven't wrapped already, we've a non-ok response, and our content is some kind of text
			if !errorWrapped && (rw.Header().Get("Content-Type") == "" || strings.Contains(rw.Header().Get("Content-Type"), "text")) {
				DebugOut.Printf("%s\n", ErrRequestError{r, fmt.Sprintf("ErrorWrapper handling %d error", rw.Code())})

				rw.Header().Del("Content-Type")

				erw := prw.NewPluggableResponseWriter()
				defer erw.Close()
				erw.SetHeader(rw.Header())
				if e.E(rw.Code(), r, erw, rw.Body.Bytes()) {
					rw.Body.ResetFromReader(erw.Body)
				}
			}

			TimingOut.Printf("ErrorWrapper took %s\n", t.Since().String())
		}
	}

	return http.HandlerFunc(fn)
}

// HandleGenericWrapper is essentially a noop for when no tempate or remote errorhandler is defined
func HandleGenericWrapper(code int, r *http.Request, rw *prw.PluggableResponseWriter, body []byte) bool {
	rw.WriteHeader(code)
	rw.Write(body)
	return true
}

// HandleTemplateWrapper wraps errors (HTTP codes >= 400) in a pretty wrapper for client presentation,
// using a template
func HandleTemplateWrapper(code int, r *http.Request, rw *prw.PluggableResponseWriter, body []byte) bool {

	te := TemplateError{
		ErrorCode:    strconv.Itoa(code),
		ErrorMessage: string(body),
	}

	rw.WriteHeader(code)
	err := ErrorTemplate.Execute(rw, te)
	if err != nil {
		ErrorOut.Println(ErrRequestError{r, fmt.Sprintf("error executing template: %s", err)})
	}
	return true

}

// HandleRemoteWrapper wraps errors (HTTP codes >= 400) in a pretty wrapper for client presentation,
// using a Worker to make a subrequest to an error-wrapping API
func HandleRemoteWrapper(code int, r *http.Request, rw *prw.PluggableResponseWriter, body []byte) bool {

	url := fmt.Sprintf("%s?ERROR_CODE=%d&ERROR_MESSAGE=%s", Conf.GetString(ConfigErrorHandlerURI), code, url.QueryEscape(string(body)))
	DebugOut.Printf("ErrorHandler GETting '%s'\n", url)
	req, _ := http.NewRequest("GET", url, nil)
	rChan := make(chan interface{}, 1)
	AddWork(&HTTPWork{
		Client:       DefaultClient,
		Request:      req,
		ResponseChan: rChan,
	})

	resp := <-rChan
	switch rt := resp.(type) {
	case error:
		ErrorOut.Println(ErrRequestError{r, fmt.Sprintf("Cascading error in RemoteErrorHandler. Demurring: %s", resp.(error))})
	case *http.Response:
		resp := resp.(*http.Response)
		defer resp.Body.Close()

		rbody, err := io.ReadAll(resp.Body)
		if err != nil {
			ErrorOut.Println(ErrRequestError{r, fmt.Sprintf("Error in RemoteErrorHandler during response body read. Demurring: %s\n", err)})
		} else {
			// It's good (badpokerface)
			// we should really verify something here first
			rw.WriteHeader(code)
			rw.Write(rbody)
			return true
		}
	default:
		ErrorOut.Println(ErrRequestError{r, fmt.Sprintf("Error in RemoteErrorHandler. Response type of '%s' unknown. Demurring.\n", rt)})
	}

	return false
}
