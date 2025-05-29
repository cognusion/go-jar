package workers

import (
	. "github.com/smartystreets/goconvey/convey"

	"testing"
	"time"
)

func init() {
	//DebugOut = log.New(os.Stderr, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
}

func TestSimpleWorkerpool(t *testing.T) {
	workChan := make(chan Work)
	rChan := make(chan interface{}, 1)

	Convey("When a simple Workerpool intializes, and when we give it work", t, func() {
		p := NewSimpleWorkerPool(workChan)
		defer p.Stop()

		So(p.Size(), ShouldEqual, -1) // stupid if{} for simple pools

		p.Metrics.Mark(1)
		workChan <- &DemoWork{rChan}

		Convey("we should get the expected response, on the return channel", func() {
			resp := <-rChan
			So(resp, ShouldHaveSameTypeAs, "Hello World")
			So(resp, ShouldEqual, "Worrrrrrrrk")
			So(p.Metrics.Count(), ShouldEqual, 1)
		})

	})
}

func TestWorkerpoolDo(t *testing.T) {
	workChan := make(chan Work)
	rChan := make(chan interface{}, 1)

	Convey("When a Workerpool intializes, the pool size is correct, and when we give it work", t, func() {
		p := NewWorkerPool(workChan, 1, 0)
		defer p.Stop()

		So(p.Size(), ShouldEqual, 1)

		p.Metrics.Mark(1)
		workChan <- &DemoWork{rChan}

		Convey("we should get the expected response, on the return channel", func() {
			resp := <-rChan
			So(resp, ShouldHaveSameTypeAs, "Hello World")
			So(resp, ShouldEqual, "Worrrrrrrrk")
		})

	})
}

func TestWorkerpoolStop(t *testing.T) {
	workChan := make(chan Work)
	//rChan := make(chan interface{}, 1)

	Convey("When a Workerpool intializes, the pool size is correct", t, func() {
		p := NewWorkerPool(workChan, 1, 0)

		So(p.Size(), ShouldEqual, 1)

		Convey("and when we stop it, the pool size is correct", func() {
			p.Stop()
			So(p.Size(), ShouldEqual, 0)
		})

	})
}

func TestWorkerpoolStopStopStop(t *testing.T) {
	workChan := make(chan Work)
	//rChan := make(chan interface{}, 1)

	Convey("When a Workerpool intializes, the pool size is correct", t, func() {
		p := NewWorkerPool(workChan, 1, 0)

		So(p.Size(), ShouldEqual, 1)

		Convey("and when we stop it, the pool size is correct", func() {
			p.Stop()
			So(p.Size(), ShouldEqual, 0)

			Convey("and when we stop it again and again, we don't block or panic", func() {
				So(p.Stop, ShouldNotPanic)
				So(p.Stop, ShouldNotPanic)
			})
		})

	})
}

func TestWorkerpoolRemove(t *testing.T) {
	workChan := make(chan Work)

	Convey("When a Workerpool intializes, the pool size is correct", t, func() {
		p := NewWorkerPool(workChan, 1, 0)
		defer p.Stop()

		So(p.Size(), ShouldEqual, 1)

		Convey("and when we ask it to remove a worker, the pool size is correct", func() {
			p.RemoveWorkers(1)
			So(p.Size(), ShouldEqual, 0)
		})

	})
}

func TestWorkerpoolAdd5(t *testing.T) {
	workChan := make(chan Work)

	Convey("When a Workerpool intializes, the pool size is correct", t, func() {
		p := NewWorkerPool(workChan, 1, 0)
		defer p.Stop()

		So(p.Size(), ShouldEqual, 1)

		Convey("and when we ask it to add five workers, the pool size is correct", func() {
			p.AddWorkers(5)
			So(p.Size(), ShouldEqual, 6)
		})

	})
}

func TestWorkerpoolCheckShrink(t *testing.T) {
	workChan := make(chan Work, 10)

	Convey("When a Workerpool intializes, the pool size is correct", t, func() {
		p := NewWorkerPool(workChan, 1, 0)
		defer p.Stop()

		So(p.Size(), ShouldEqual, 1)

		Convey("and when we ask it to add five workers, the pool size is correct", func() {
			p.AddWorkers(5)
			So(p.Size(), ShouldEqual, 6)

			Convey("and when we run CheckAndAdjust(), we should shed the five added workers", func() {
				p.checkAndAdjust()
				So(p.Size(), ShouldEqual, 1)
			})
		})

	})
}

func TestWorkerpoolCheckWontShrink(t *testing.T) {
	workChan := make(chan Work, 10)

	Convey("When a Workerpool intializes, the pool size is correct", t, func() {
		p := NewWorkerPool(workChan, 1, 0)
		defer p.Stop()

		So(p.Size(), ShouldEqual, 1)

		Convey("and when we ask it to add five workers, the pool size is correct", func() {
			p.AddWorkers(5)
			So(p.Size(), ShouldEqual, 6)

			Convey("and when we run CheckAndAdjust()", func() {
				p.checkAndAdjust()
				So(p.Size(), ShouldEqual, 1)
				Convey("and then we bump up the min and run CheckAndAdjust(), nothing should change", func() {
					p.Min(5)
					p.checkAndAdjust()
					So(p.Size(), ShouldEqual, 1)
				})
			})
		})

	})
}

func TestWorkerpoolCheckGrow(t *testing.T) {
	workChan := make(chan Work, 10)
	rChan := make(chan interface{}, 1000)

	Convey("When a Workerpool intializes, the pool size is correct", t, func() {
		p := NewWorkerPool(workChan, 1, 0)
		defer p.Stop()

		So(p.Size(), ShouldEqual, 1)

		go func() {
			for i := 0; i < 1000; i++ {
				if p.Size() <= 0 {
					break
				}
				workChan <- &SlowDemoWork{rChan}
				<-p.adjustLock
				p.Metrics.Mark(1)
				p.adjustLock <- true
			}
		}()
		Convey("and when we ask it to add five workers, the pool size is correct", func() {
			p.AddWorkers(5)
			So(p.Size(), ShouldEqual, 6)

			Convey("and when we run CheckAndAdjust(), queue size (10) more workers should be added", func() {
				time.Sleep(1 * time.Second)
				p.checkAndAdjust()
				So(p.Size(), ShouldEqual, 16)
			})
		})

	})
}

func TestWorkerpoolCheckWork(t *testing.T) {
	workChan := make(chan Work, 10)
	rChan := make(chan interface{}, 1000)

	Convey("When a Workerpool intializes, the pool size is correct", t, func() {
		p := NewWorkerPool(workChan, 0, 0)
		defer p.Stop()

		So(p.Size(), ShouldEqual, 0)

		p.Metrics.Mark(1)
		workChan <- &SlowDemoWork{rChan}

		Convey("and when we add some work, with no workers running it should be visible", func() {
			So(p.Work(), ShouldEqual, 1)

			Convey("and when we run CheckAndAdjust(), more workers should be added", func() {
				p.checkAndAdjust()
				time.Sleep(1 * time.Second)
				So(p.Size(), ShouldEqual, 1)
			})
		})

	})
}

func TestWorkerpoolCheckGrowMax(t *testing.T) {
	workChan := make(chan Work, 10)
	rChan := make(chan interface{}, 1000)

	Convey("When a Workerpool intializes, and the Max is set to 5, the pool size is correct", t, func() {
		p := NewWorkerPool(workChan, 1, 0)
		defer p.Stop()

		So(p.Size(), ShouldEqual, 1)

		p.Max(5)
		go func() {
			for i := 0; i < 1000; i++ {
				workChan <- &SlowDemoWork{rChan}
				<-p.adjustLock
				p.Metrics.Mark(1)
				p.adjustLock <- true
			}
		}()
		Convey("and when we run CheckAndAdjust(), queue size (10) more workers should be added, but in reality no more than our max (5)", func() {
			time.Sleep(1 * time.Second)
			p.checkAndAdjust()
			So(p.Size(), ShouldEqual, 5)
		})
	})
}

func TestWorkerpoolCheckGrowBusyish(t *testing.T) {
	t.Skip("Slow...\n")

	workChan := make(chan Work, 10)

	Convey("When a Workerpool intializes, the pool size is correct", t, func() {
		p := NewWorkerPool(workChan, 1, 0)
		defer p.Stop()

		So(p.Size(), ShouldEqual, 1)

		p.Metrics.Mark(50)
		time.Sleep(3 * time.Second)

		Convey("and when we run add 50 pieces of work, and then run CheckAndAdjust(), bump the workers up to up to, which is 19", func() {
			p.checkAndAdjust()
			So(p.Size(), ShouldEqual, 19)
		})
	})
}

func TestWorkerpoolCheckShrinkBusyish(t *testing.T) {
	t.Skip("Slow...\n")

	workChan := make(chan Work, 10)

	Convey("When a Workerpool intializes, the pool size is correct", t, func() {
		p := NewWorkerPool(workChan, 50, 0)
		defer p.Stop()

		for i := 0; i < 30; i++ {
			<-p.adjustLock
			p.Metrics.Mark(1)
			p.adjustLock <- true
			time.Sleep(100 * time.Millisecond)
		}

		p.Min(1)

		So(p.Size(), ShouldEqual, 50)

		Convey("and when we run CheckAndAdjust(), we slam it down to the workload, which is 12", func() {
			time.Sleep(3 * time.Second)
			p.checkAndAdjust()
			So(p.Size(), ShouldEqual, 12)
		})
	})
}

func TestWorkerpoolCheckGrowAutomatic(t *testing.T) {
	t.Skip("Slow....\n")

	workChan := make(chan Work, 10)
	rChan := make(chan interface{}, 1000)

	Convey("When a Workerpool intializes, the pool size is correct", t, func() {
		p := NewWorkerPool(workChan, 1, 2*time.Second)
		defer p.Stop()

		So(p.Size(), ShouldEqual, 1)

		go func() {
			for i := 0; i < 1000; i++ {
				workChan <- &SlowDemoWork{rChan}
				<-p.adjustLock
				p.Metrics.Mark(1)
				p.adjustLock <- true
			}
		}()
		Convey("and when we wait a coupleseconds queue size (10) more workers should be added", func() {
			time.Sleep(3 * time.Second)
			So(p.Size(), ShouldEqual, 11)
		})

	})
}

func addWork(pool *WorkerPool, work Work) {
	//pool.Metrics.Mark(1)
	pool.WorkChan <- work
}

func Benchmark_OldWorkerPool(b *testing.B) {
	workChan := make(chan Work)
	rChan := make(chan interface{})

	p := NewWorkerPool(workChan, 100, 0)
	defer p.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Yes, this is super naive and doesn't test the overall async performance.
		// I'm benching for allocs more than time.
		addWork(p, &DemoWork{rChan})
		<-rChan
	}
}

func Benchmark_SimpleWorkerPool(b *testing.B) {
	workChan := make(chan Work)
	rChan := make(chan interface{})

	p := NewSimpleWorkerPool(workChan)
	defer p.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Yes, this is super naive and doesn't test the overall async performance.
		// I'm benching for allocs more than time.
		addWork(p, &DemoWork{rChan})
		<-rChan
	}
}
