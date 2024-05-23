package jar

import (
	"github.com/cognusion/go-jar/workers"
)

var (
	// Workers are a pool of workers
	Workers *workers.WorkerPool
	// AddWork queues up some work for workers
	AddWork func(workers.Work)
)

func init() {

	InitFuncs.Add(func() {
		workers.DebugOut = DebugOut
		workers.ErrorOut = ErrorOut

		if Workers == nil {
			// Since InitFuncs may be called multiple times, we don't want to orphan these
			workChan := make(chan workers.Work)

			Workers = workers.NewSimpleWorkerPool(workChan)

			AddWork = func(work workers.Work) {
				Workers.Metrics.Mark(1)
				workChan <- work
			}
		}
	})
}
