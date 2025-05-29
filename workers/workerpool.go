package workers

import (
	"github.com/rcrowley/go-metrics"

	"io"
	"log"
)

var (
	// DebugOut is a log.Logger for debug messages
	DebugOut = log.New(io.Discard, "[DEBUG] ", 0)
	// ErrorOut is a log.Logger for error messages
	ErrorOut = log.New(io.Discard, "", 0)
)

// WorkerPool is a simple mechanation to manage and accomplish Work.
type WorkerPool struct {
	// WorkChan is where the work goes
	WorkChan chan Work
	// Stop permanently stops the pool after current work is done.
	// WorkChan is not closed, to prevent errant reads
	Stop func()

	Metrics metrics.Meter

	// killChan will close when all the Workers need to exit
	killChan chan struct{}
}

// NewSimpleWorkerPool returns a functioning WorkerPool bound to WorkChan, that does not try keep Workers running. It is strongly recommended that
// WorkChan be unbounded unless you demonstrably need to bound it.
func NewSimpleWorkerPool(WorkChan chan Work) *WorkerPool {
	p := &WorkerPool{
		WorkChan: WorkChan,
		Metrics:  metrics.NewMeter(),
		killChan: make(chan struct{}),
	}

	p.Stop = func() {
		p.Stop = func() {}
		p.Metrics.Stop()
		close(p.killChan)
	}

	go p.simplePool() // make it so

	return p
}

// AddWork adds Work and updates metrics. Will panic if WorkChan is closed.
func (p *WorkerPool) AddWork(work Work) {
	p.WorkChan <- work
	p.Metrics.Mark(1)
}

// Work returns the quantity of Work in the work channel
func (p *WorkerPool) Work() int {
	return len(p.WorkChan)
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
