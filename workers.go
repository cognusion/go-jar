package jar

import (
	"github.com/cognusion/go-jar/workers"
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
