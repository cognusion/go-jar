package jar

import (
	"github.com/eapache/go-resiliency/retrier"

	"fmt"
	"net/http"
	"time"
)

// HTTPWork is a generic Work that can make HTTP requests
type HTTPWork struct {
	Client       *http.Client
	Request      *http.Request
	ResponseChan chan interface{}
	// RetryCount is the number of times to retry Request if there is an error
	RetryCount int
	//RetryInterval is the duration between retries
	RetryInterval time.Duration
	//RetryHTTPErrors, if set, classifies HTTP responses >= 500 as errors for retry purposes
	RetryHTTPErrors bool
}

// Work is called to do work
func (h *HTTPWork) Work() interface{} {

	var ret *http.Response
	try := func() error {
		resp, err := h.Client.Do(h.Request)
		if err != nil {
			return err
		}
		if h.RetryHTTPErrors && resp.StatusCode >= 500 {
			return fmt.Errorf("%d %s", resp.StatusCode, resp.Status)
		}
		ret = resp
		return nil
	}

	r := retrier.New(retrier.ConstantBackoff(h.RetryCount, h.RetryInterval), nil)

	err := r.Run(try)
	if err != nil {
		return err
	}
	return ret
}

// Return is called response with results
func (h *HTTPWork) Return(rthing interface{}) {
	h.ResponseChan <- rthing
}
