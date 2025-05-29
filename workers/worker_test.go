package workers

import (
	. "github.com/smartystreets/goconvey/convey"

	"testing"
	"time"
)

func TestWorkerPanic(t *testing.T) {
	workChan := make(chan Work)
	quitChan := make(chan bool)
	killChan := make(chan struct{})
	rChan := make(chan interface{}, 1)
	defer close(killChan)

	Convey("When a Worker intializes and gets work", t, func() {
		r := Worker{
			WorkChan: workChan,
			QuitChan: quitChan,
			KillChan: killChan,
		}
		r.Do()

		So(r.self.Load(), ShouldBeTrue)

		workChan <- &PanicWork{rChan}

		Convey("and panics, we should get the expected panic response, on the return channel", func() {
			resp := <-rChan
			So(resp, ShouldHaveSameTypeAs, WorkError{})
			weResp := resp.(WorkError)
			So(weResp.Error(), ShouldEqual, "Work generated panic: OH NO!!!!!!")
		})

	})
}

func TestWorkerDo(t *testing.T) {
	workChan := make(chan Work)
	quitChan := make(chan bool)
	killChan := make(chan struct{})
	rChan := make(chan interface{}, 1)
	defer close(killChan)

	Convey("When a Worker intializes and gets work", t, func() {
		r := Worker{
			WorkChan: workChan,
			QuitChan: quitChan,
			KillChan: killChan,
		}
		r.Do()

		So(r.self.Load(), ShouldBeTrue)

		workChan <- &DemoWork{rChan}

		Convey("we should get the expected response, on the return channel", func() {
			resp := <-rChan
			So(resp, ShouldHaveSameTypeAs, "Hello World")
			So(resp, ShouldEqual, "Worrrrrrrrk")
		})

	})
}

func TestWorkerDoOnce(t *testing.T) {
	rChan := make(chan interface{}, 1)

	Convey("When a Worker is told to do some work once", t, func() {
		r := Worker{}
		r.DoOnce(&DemoWork{rChan})

		Convey("we should get the expected response, on the return channel", func() {
			resp := <-rChan
			So(resp, ShouldHaveSameTypeAs, "Hello World")
			So(resp, ShouldEqual, "Worrrrrrrrk")
		})

	})
}

func TestWorkerQuit(t *testing.T) {
	workChan := make(chan Work)
	quitChan := make(chan bool)
	killChan := make(chan struct{})
	//rChan := make(chan interface{}, 1)
	defer close(killChan)

	Convey("When a Worker intializes and is told to quit", t, func() {
		r := Worker{
			WorkChan: workChan,
			QuitChan: quitChan,
			KillChan: killChan,
		}
		r.Do()
		So(r.self.Load(), ShouldBeTrue)

		quitChan <- true
		time.Sleep(1 * time.Millisecond)
		Convey("it's gone", func() {
			So(r.self.Load(), ShouldBeFalse)
		})

	})
}

func TestWorkerKill(t *testing.T) {
	workChan := make(chan Work)
	quitChan := make(chan bool)
	killChan := make(chan struct{})
	//rChan := make(chan interface{}, 1)

	Convey("When a Worker intializes and is killed", t, func() {
		r := Worker{
			WorkChan: workChan,
			QuitChan: quitChan,
			KillChan: killChan,
		}
		r.Do()

		close(killChan)

		// Race
		time.Sleep(100 * time.Millisecond)

		Convey("it's gone", func() {
			So(r.self.Load(), ShouldBeFalse)
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
