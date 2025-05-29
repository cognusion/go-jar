package workers

import (
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"testing"
)

func init() {
	//DebugOut = log.New(os.Stderr, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
}

func TestSimpleWorkerpool(t *testing.T) {
	workChan := make(chan Work)
	rChan := make(chan any)

	Convey("When a simple WorkerPool intializes, and when we give it work", t, func() {
		p := NewSimpleWorkerPool(workChan)
		defer p.Stop()

		p.AddWork(&DemoWork{rChan})

		Convey("we should get the expected response on the return channel, and the instrumentations should be accurate", func() {
			resp := <-rChan
			So(resp, ShouldHaveSameTypeAs, "Hello World")
			So(resp, ShouldEqual, "Worrrrrrrrk")
			So(p.Metrics.Count(), ShouldEqual, 1)
			So(p.Work(), ShouldEqual, 0)
		})
	})
}

func TestWorkerPanic(t *testing.T) {
	rChan := make(chan interface{})

	Convey("When a Worker intializes and gets work", t, func() {
		go workIt(&PanicWork{rChan})

		Convey("and panics, we should get the expected panic response, on the return channel", func() {
			resp := <-rChan
			So(resp, ShouldHaveSameTypeAs, WorkError{})
			weResp := resp.(WorkError)
			So(weResp.Error(), ShouldEqual, "Work generated panic: OH NO!!!!!!")
		})

	})
}

func TestWorkerDo(t *testing.T) {
	rChan := make(chan interface{}, 1)

	Convey("When a Worker intializes and gets work", t, func() {
		go workIt(&DemoWork{rChan})

		Convey("we should get the expected response, on the return channel", func() {
			resp := <-rChan
			So(resp, ShouldHaveSameTypeAs, "Hello World")
			So(resp, ShouldEqual, "Worrrrrrrrk")
		})

	})
}

type SlowDemoWork struct {
	ResponseChan chan interface{}
}

func (w *SlowDemoWork) Work() interface{} {
	time.Sleep(10 * time.Second)
	return "Worrrrrrrrk"
}

func (w *SlowDemoWork) Return(work interface{}) {
	w.ResponseChan <- work
}

type DemoWork struct {
	ResponseChan chan interface{}
}

func (w *DemoWork) Work() interface{} {
	return "Worrrrrrrrk"
}

func (w *DemoWork) Return(work interface{}) {
	w.ResponseChan <- work
}

type PanicWork struct {
	ResponseChan chan interface{}
}

func (w *PanicWork) Work() interface{} {
	panic("OH NO!!!!!!")
}

func (w *PanicWork) Return(work interface{}) {
	w.ResponseChan <- work
}

func Benchmark_SimpleWorkerPool(b *testing.B) {
	workChan := make(chan Work)
	rChan := make(chan any)

	p := NewSimpleWorkerPool(workChan)
	defer p.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Yes, this is super naive and doesn't test the overall async performance.
		// I'm benching for allocs more than time.
		p.AddWork(&DemoWork{rChan})
		<-rChan
	}
}
