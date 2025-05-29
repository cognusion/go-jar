package workers

import "fmt"

// Work is an interface to allow the abstraction of Work and Return,
// enabling generic Workers doing blind Work
type Work interface {
	Work() any
	Return(any)
}

// WorkError is sent to Work.Return() if the Work generates a panic
type WorkError struct {
	Messages string
}

func (w *WorkError) Error() string {
	return w.Messages
}

// workIt performs work, handling panics as possible.
func workIt(work Work) {
	defer recovery(work)
	work.Return(work.Work())
}

// recovery is used to recover from panic's caused by shoddy Work
func recovery(panicwork Work) {
	if r := recover(); r != nil {
		DebugOut.Printf("Worker panic'd: %s\n", r)
		ErrorOut.Printf("Worker panic'd: %s\n", r)

		if panicwork != nil {
			panicwork.Return(WorkError{fmt.Sprintf("Work generated panic: %s", r)})
		}
	}
}
