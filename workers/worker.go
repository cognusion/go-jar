package workers

import (
	"sync/atomic"
)

// Worker is a simple primitive construct that listens on WorkChan for Work to do,
// Might hear a "true" on QuitChan if it is underworked,
// Might see a closed KillChan if it's time to leave expeditiously
type Worker struct {
	// WorkChan is where the work comes from
	WorkChan chan Work
	// QuitChan will get some bools sent to it when the Worker pool needs to shrink
	QuitChan chan bool
	// KillChan will close when all the Workers need to exit
	KillChan chan struct{}

	self atomic.Value
}

// DoOnce asks the Worker to do one unit of Work and then end
func (w *Worker) DoOnce(work Work) {
	workIt(work)
}

// Do forks off a Workerloop that listens for Work, quits, or kills
func (w *Worker) Do() {
	w.self.Store(true)

	go func() {
		Close := func() { w.self.Store(false) }
		defer Close()

		for {
			select {
			case work := <-w.WorkChan:
				DebugOut.Printf("Work: %T %+v \n", work, work)
				workIt(work)
			case b := <-w.QuitChan:
				if b {
					return
				}
			case <-w.KillChan:
				return
			}
		}
	}()
}
