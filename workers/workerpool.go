package workers

import (
	"github.com/rcrowley/go-metrics"

	"io"
	"log"
	"math"
	"sync/atomic"
	"time"
)

var (
	// DebugOut is a log.Logger for debug messages
	DebugOut = log.New(io.Discard, "[DEBUG] ", 0)
	// ErrorOut is a log.Logger for error messages
	ErrorOut = log.New(io.Discard, "", 0)
)

// WorkerPool was an overly-complicated mechanation to allow arbitrary work to be accomplished by an arbitrary worker.
// If you obtain your WorkerPool using NewSimpleWorkerPool, most of the over-complicated mechanation is ignored.
// If you obtain your WorkerPool using NewWorkerPool, the following is deprecated, and applies:
//
// WorkerPool is an overly-complicated mechanation to allow arbitrary work to be accomplished by an arbitrary worker,
// which will then return arbitrary results onto an arbitrary channel, while allowing for the evidence-driven growing or
// shrinking of the pool of available workers based on the fillyness of the WorkChan, which should be buffered and of
// an appropriate size. If that hasn't turned you off yet, carry on.
type WorkerPool struct {
	// WorkChan is where the work goes
	WorkChan chan Work
	// Stop permanently stops the pool after current work is done.
	// WorkChan is not closed, to prevent errant reads
	Stop func()

	Metrics metrics.Meter

	// QuitChan will get some bools sent to it when the Worker pool needs to shrink
	quitChan chan bool
	// KillChan will close when all the Workers need to exit
	killChan chan struct{}
	// Size is the eventually-consistent number of workers in the pool
	size int64
	// minpool is the minimum size of the pool
	minpool int64
	// maxpool is the maximum size of the pool
	maxpool int64

	adjustLock     chan bool
	adjustInterval time.Duration
	simple         bool
}

// NewSimpleWorkerPool returns a functioning WorkerPool bound to WorkChan, that does not try keep Workers running. It is strongly recommended that
// WorkChan be unbounded unless you demonstrably need to bound it.
func NewSimpleWorkerPool(WorkChan chan Work) *WorkerPool {
	p := &WorkerPool{
		WorkChan: WorkChan,
		Metrics:  metrics.NewMeter(),
		killChan: make(chan struct{}),
		simple:   true,
	}

	p.Stop = func() {
		p.Stop = func() {}
		p.Metrics.Stop()
		close(p.killChan)
	}

	go p.simplePool() // make it so

	return p
}

// NewWorkerPool returns a functioning WorkerPool bound to WorkChan, with an initial pool size of initialSize, and if autoAdjustInterval > 0, then
// it will run the CheckAndAdjust() every that often.
// NOTE: If your WorkChan is unbuffered (no size during make(), autoAdjust will not run, nor will calling CheckAndAdjust() result in changes. The channel capacity
// and usage is key to this. It is recommended that the buffer size be around anticipated burst size for work
//
// Deprecated: NewWorkerPool has been replaced by NewSimpleWorkerPool. In v2 this version will be removed completely in lieu of the "simple" variety.
func NewWorkerPool(WorkChan chan Work, initialSize int, autoAdjustInterval time.Duration) *WorkerPool {
	p := &WorkerPool{
		WorkChan:       WorkChan,
		Metrics:        metrics.NewMeter(),
		quitChan:       make(chan bool, initialSize),
		killChan:       make(chan struct{}),
		size:           int64(0),
		minpool:        int64(2),
		maxpool:        int64(0),
		adjustLock:     make(chan bool, 1),
		adjustInterval: autoAdjustInterval,
	}

	// Edge case
	if initialSize == 1 {
		p.minpool = int64(1)
	}

	// Prime the lock
	p.adjustLock <- true

	p.Stop = func() {
		p.Stop = func() {}
		// TODO let workers flag their exits

		<-p.adjustLock
		// we have lock
		// Make sure we unlock it
		defer func() { p.adjustLock <- true }()

		p.Metrics.Stop()
		atomic.StoreInt64(&p.size, int64(0))
		close(p.killChan)
	}

	if initialSize > 0 {
		p.AddWorkers(int64(initialSize))
	}

	if autoAdjustInterval > 0 {
		if cap(p.WorkChan) > 0 {
			// Channel is buffered, we may proceed
			go func() {
				ticker := time.NewTicker(autoAdjustInterval)
				defer ticker.Stop()

				for {
					select {
					case <-p.killChan:
						// Kill signalled
						return
					case <-ticker.C:
						p.CheckAndAdjust()
					}
				}
			}()
		}
	}

	return p
}

// Work returns the quantity of Work in the work channel
func (p *WorkerPool) Work() int {
	return len(p.WorkChan)
}

// Min sets the minimum number of workers
//
// Deprecated: Min will be removed in v2
func (p *WorkerPool) Min(min int) {
	if p.simple {
		return
	}
	atomic.StoreInt64(&p.minpool, int64(min))
}

// Max sets the maximum number of workers
//
// Deprecated: Max will be removed in v2
func (p *WorkerPool) Max(max int) {
	if p.simple {
		return
	}
	atomic.StoreInt64(&p.maxpool, int64(max))
}

// CheckAndAdjust asynchronously triggers the process to possibly resize the pool.
// While a resize process is running, subsequent processors will silently exit
//
// Deprecated: CheckAndAdjust will be removed in v2
func (p *WorkerPool) CheckAndAdjust() {
	if p.simple {
		return
	}
	go p.checkAndAdjust()
}

// checkAndAdjust is the workhorse for CheckAndAdjust.
func (p *WorkerPool) checkAndAdjust() {
	select {
	case <-p.adjustLock:
		// we have lock
		// Make sure we unlock it
		defer func() { p.adjustLock <- true }()
	default:
		// we do not, someone else is working here, skip on
		DebugOut.Print("checkAndAdjust detects another run, skipping\n")
		return
	}

	var rate float64
	if p.adjustInterval < 5*time.Minute {
		rate = p.Metrics.Rate1()
	} else if p.adjustInterval < 15*time.Minute {
		rate = p.Metrics.Rate5()
	} else {
		rate = p.Metrics.Rate15()
	}

	var (
		ratefix  = int(math.Ceil(rate))
		poolsize = p.size
		qsize    = cap(p.WorkChan)
		qcount   = len(p.WorkChan)
	)
	DebugOut.Printf("checkAndAdjust: Rate: %.4f (%d) Qsize: %d Qcount: %d Poolsize: %d\n", rate, ratefix, qsize, qcount, poolsize)

	if p.minpool > 0 && p.size > p.minpool && qcount == 0 {
		// We allow shrinking
		if ratefix == 0 && poolsize > p.minpool {
			// No work, pool is way too big. Slam it down
			diff := poolsize - p.minpool
			DebugOut.Print("\tNo work, pool is way too big. Slam it down\n")
			p.RemoveWorkers(diff)
			return
		} else if ratefix == 0 {
			// No work, but the pool is small enough. We're done
			return
		}
		if poolsize > int64(ratefix*2) {
			// Pool is over twice the rate, cut it back
			diff := poolsize - int64(ratefix*2)
			if poolsize-diff < p.minpool {
				diff = poolsize - p.minpool
			}
			DebugOut.Printf("\tPool is over twice the rate, cut it back\n")
			p.RemoveWorkers(diff)
			return
		}
	}

	if qsize > ratefix && qcount > 0 {
		// Stuff sitting in the queue? Waaaat!?
		if p.maxpool > 0 && poolsize+int64(qcount) > p.maxpool {
			// We have a maxpool, and we'd bust it, so add up to it
			qcount = int(p.maxpool - poolsize)
		}
		DebugOut.Print("\tStuff sitting in the queue? Waaaat!?\n")
		p.AddWorkers(int64(qcount))
		return
	}

	if int64(ratefix) > poolsize {
		// More work than workers
		diff := int64(ratefix) - poolsize
		if p.maxpool > 0 && poolsize+(diff*2) > p.maxpool {
			// We have a maxpool, and we'd bust it, so add up to it
			diff = p.maxpool - poolsize
		}
		DebugOut.Print("\tMore work than workers!\n")
		p.AddWorkers(diff * 2) // we add twice as many as we need, and will trickle them off later
		return
	}

}

// AddWorkers adds the specified number of workers
//
// Deprecated: AddWorkers will be removed in v2
func (p *WorkerPool) AddWorkers(number int64) {
	if p.simple {
		return
	}

	DebugOut.Printf("\tAdding %d workers\n", number)
	for i := int64(0); i < number; i++ {
		w := Worker{
			WorkChan: p.WorkChan,
			QuitChan: p.quitChan,
			KillChan: p.killChan,
		}
		w.Do()
	}
	atomic.AddInt64(&p.size, number)
}

// RemoveWorkers removes the specified number of workers, or the number running
//
// Deprecated: RemoveWorkers will be removed in v2
func (p *WorkerPool) RemoveWorkers(number int64) {
	if p.simple {
		return
	}

	if number > p.Size() {
		number = p.Size()
	}
	DebugOut.Printf("\tRemoving %d workers\n", number)
	for i := int64(0); i < number; i++ {
		// We're not blocking here, because busy.
		// Maybe we should not return until the quits are done?
		go func() {
			p.quitChan <- true
		}()
	}
	atomic.AddInt64(&p.size, -1*number)
}

// Size returns the eventually-consistent number of workers in the pool. Returns -1 for simple pools
//
// Deprecated: Size will be removed in v2
func (p *WorkerPool) Size() int64 {
	if p.simple {
		return -1 // This is likely stupid, but ... -_o_-
	}

	return atomic.LoadInt64(&p.size)
}

// simplePool listens for work, and doles it out to a new Worker.
func (p *WorkerPool) simplePool() {
	for {
		select {
		case work := <-p.WorkChan:
			DebugOut.Printf("Work: %T %+v \n", work, work)

			go workIt(work)

		case <-p.killChan:
			return
		}
	}
}
