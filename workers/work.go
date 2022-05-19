package workers

// Work is an interface to allow the abstraction of Work and Return,
// enabling generic Workers doing blind Work
type Work interface {
	Work() interface{}
	Return(interface{})
}

// WorkError is sent to Work.Return() if the Work generates a panic
type WorkError struct {
	Messages string
}

func (w *WorkError) Error() string {
	return w.Messages
}
