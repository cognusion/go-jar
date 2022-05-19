package jar

import (
	"github.com/eapache/go-resiliency/retrier"

	"github.com/cognusion/go-jar/workers"

	"fmt"
	"net/http"
	"time"
)

// Constants for configuration key strings
const (
	ConfigWorkersInitialPoolSize = ConfigKey("workers.initialpoolsize")
	ConfigWorkersMaxPoolSize     = ConfigKey("workers.maxpoolsize")
	ConfigWorkersMinPoolSize     = ConfigKey("workers.minpoolsize")
	ConfigWorkersQueueSize       = ConfigKey("workers.queuesize")
	ConfigWorkersResizeInterval  = ConfigKey("workers.resizeinterval")
)

var (
	// Workers are a pool of workers
	Workers *workers.WorkerPool
	// AddWork queues up some work for workers
	AddWork func(workers.Work)
)

func init() {
	ConfigAdditions[ConfigWorkersQueueSize] = 100
	ConfigAdditions[ConfigWorkersInitialPoolSize] = 10
	ConfigAdditions[ConfigWorkersMinPoolSize] = 2
	ConfigAdditions[ConfigWorkersMaxPoolSize] = 0
	ConfigAdditions[ConfigWorkersResizeInterval] = "30s"

	InitFuncs.Add(func() {
		workers.DebugOut = DebugOut
		workers.ErrorOut = ErrorOut

		if Workers == nil {
			// Since InitFuncs may be called multiple times, we don't want to orphan these
			workChan := make(chan workers.Work, Conf.GetInt(ConfigWorkersQueueSize))
			AddWork = func(work workers.Work) {
				Workers.Metrics.Mark(1)
				workChan <- work
			}
			Workers = workers.NewWorkerPool(workChan, Conf.GetInt(ConfigWorkersInitialPoolSize), Conf.GetDuration(ConfigWorkersResizeInterval))
			Workers.Min(Conf.GetInt(ConfigWorkersMinPoolSize))
			Workers.Max(Conf.GetInt(ConfigWorkersMaxPoolSize))
		}
	})
}

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
