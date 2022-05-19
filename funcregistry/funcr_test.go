package funcregistry

import (
	. "github.com/smartystreets/goconvey/convey"

	"testing"
	"time"
)

func TestFuncr(t *testing.T) {

	fChan := make(chan bool, 1)

	fr := NewFuncRegistry(false)
	defer fr.Close()

	f := func() {
		fChan <- true
	}

	Convey("When a new FuncRegistry is created, and 1000 functions are added to it", t, func() {
		for i := 0; i < 1000; i++ {
			fr.Add(f)
		}
		fr.Add(func() {
			fChan <- false
		})

		go fr.Call()

		Convey("after Call()ing, all of the results are accounted for", func() {
			returnCount := 0
		OUT:
			for {
				select {
				case b := <-fChan:
					if b {
						returnCount++
					} else {
						break OUT
					}
				}
			}

			So(returnCount, ShouldEqual, 1000)
		})
	})
}

func TestFuncrCallOnce(t *testing.T) {

	fChan := make(chan bool, 1)

	fr := NewFuncRegistry(true)
	defer fr.Close()

	f := func() {
		fChan <- true
	}

	Convey("When a new FuncRegistry is created with \"call once\" set, and 1000 functions are added to it", t, func() {
		for i := 0; i < 1000; i++ {
			fr.Add(f)
		}
		fr.Add(func() {
			fChan <- false
		})

		go fr.Call()
		Convey("after Call()ing, all of the results are accounted for", func() {
			returnCount := 0
		OUT:
			for {
				select {
				case b := <-fChan:
					if b {
						returnCount++
					} else {
						break OUT
					}
				}
			}

			So(returnCount, ShouldEqual, 1000)

			Convey("and after Call()ing a second time, waiting results in a timeout, because it won't fire again", func() {
				go fr.Call()

				timeout := time.After(time.Second)
				select {
				case _ = <-fChan:
					t.Errorf("Second Call() executed!\n")
				case <-timeout:
					// Success
				}
			})
		})
	})
}

func TestFuncrAddOnceCall(t *testing.T) {

	fChan := make(chan bool, 1)

	fr := NewFuncRegistry(false)
	defer fr.Close()

	f := func() {
		fChan <- true
	}

	Convey("When a new FuncRegistry is created, and 900 functions are AddOnce()d, and 100 functions are Add()d", t, func() {

		for i := 0; i < 900; i++ {
			fr.AddOnce(f)
		}
		for i := 0; i < 100; i++ {
			fr.Add(f)
		}

		fr.AddOnce(func() {
			fChan <- false
		})

		go fr.Call()
		Convey("after Call()ing once, all 1000 of the results are accounted for", func() {
			returnCount := 0
		OUT:
			for {
				select {
				case b := <-fChan:
					if b {
						returnCount++
					} else {
						break OUT
					}
				}
			}

			So(returnCount, ShouldEqual, 1000)

			Convey("and after Call()ing a second time, only the 100 Add()s are accounted for", func() {
				returnSecondCount := 0
				fr.AddOnce(func() {
					fChan <- false
				})

				go fr.Call()

				timeout := time.After(time.Second)
			OUT2:
				for {
					select {
					case b := <-fChan:
						if b {
							returnSecondCount++
						} else {
							break OUT2
						}
					case <-timeout:
						t.Error("second Call() timed out!!")
					}
				}
				So(returnSecondCount, ShouldEqual, 100)
			})
		})
	})
}

func TestFuncrCallTwice(t *testing.T) {

	fChan := make(chan bool, 1)

	fr := NewFuncRegistry(false)
	defer fr.Close()

	f := func() {
		fChan <- true
	}

	Convey("When a new FuncRegistry is created with set, and 1000 functions are added to it", t, func() {
		for i := 0; i < 1000; i++ {
			fr.Add(f)
		}
		fr.Add(func() {
			fChan <- false
		})

		go fr.Call()

		Convey("after Call()ing, all of the results are accounted for", func() {
			returnCount := 0
		OUT:
			for {
				select {
				case b := <-fChan:
					if b {
						returnCount++
					} else {
						break OUT
					}
				}
			}

			So(returnCount, ShouldEqual, 1000)

			go fr.Call()
			Convey("after Call()ing a second time, all of the results are accounted for", func() {

			OUT2:
				for {
					select {
					case b := <-fChan:
						if b {
							returnCount++
						} else {
							break OUT2
						}
					}
				}

				So(returnCount, ShouldEqual, 2000)

			})
		})
	})
}
