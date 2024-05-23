package workers

import (
	"fmt"
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
	w.workIt(work)
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
				w.workIt(work)
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

func (w *Worker) workIt(work Work) {
	defer w.recovery(work)
	work.Return(work.Work())
}

// recovery is used internally to recover from panic's caused by shoddy Work
func (w *Worker) recovery(panicwork Work) {
	if r := recover(); r != nil {
		DebugOut.Printf("Worker panic'd: %s\n", r)
		ErrorOut.Printf("Worker panic'd: %s\n", r)

		if panicwork != nil {
			panicwork.Return(WorkError{fmt.Sprintf("Work generated panic: %s", r)})
		}
	}
}
