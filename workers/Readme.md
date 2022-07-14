

# workers
`import "github.com/cognusion/go-jar/workers"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>



## <a name="pkg-index">Index</a>
* [Variables](#pkg-variables)
* [type Work](#Work)
* [type WorkError](#WorkError)
  * [func (w *WorkError) Error() string](#WorkError.Error)
* [type Worker](#Worker)
  * [func (w *Worker) Do()](#Worker.Do)
* [type WorkerPool](#WorkerPool)
  * [func NewWorkerPool(WorkChan chan Work, initialSize int, autoAdjustInterval time.Duration) *WorkerPool](#NewWorkerPool)
  * [func (p *WorkerPool) AddWorkers(number int64)](#WorkerPool.AddWorkers)
  * [func (p *WorkerPool) CheckAndAdjust()](#WorkerPool.CheckAndAdjust)
  * [func (p *WorkerPool) Max(max int)](#WorkerPool.Max)
  * [func (p *WorkerPool) Min(min int)](#WorkerPool.Min)
  * [func (p *WorkerPool) RemoveWorkers(number int64)](#WorkerPool.RemoveWorkers)
  * [func (p *WorkerPool) Size() int64](#WorkerPool.Size)
  * [func (p *WorkerPool) Work() int](#WorkerPool.Work)


#### <a name="pkg-files">Package files</a>
[work.go](https://github.com/cognusion/go-jar/tree/master/workers/work.go) [worker.go](https://github.com/cognusion/go-jar/tree/master/workers/worker.go) [workerpool.go](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go)



## <a name="pkg-variables">Variables</a>
``` go
var (
    // DebugOut is a log.Logger for debug messages
    DebugOut = log.New(io.Discard, "[DEBUG] ", 0)
    // ErrorOut is a log.Logger for error messages
    ErrorOut = log.New(io.Discard, "", 0)
)
```



## <a name="Work">type</a> [Work](https://github.com/cognusion/go-jar/tree/master/workers/work.go?s=131:195#L5)
``` go
type Work interface {
    Work() interface{}
    Return(interface{})
}
```
Work is an interface to allow the abstraction of Work and Return,
enabling generic Workers doing blind Work










## <a name="WorkError">type</a> [WorkError](https://github.com/cognusion/go-jar/tree/master/workers/work.go?s=265:307#L11)
``` go
type WorkError struct {
    Messages string
}

```
WorkError is sent to Work.Return() if the Work generates a panic










### <a name="WorkError.Error">func</a> (\*WorkError) [Error](https://github.com/cognusion/go-jar/tree/master/workers/work.go?s=309:343#L15)
``` go
func (w *WorkError) Error() string
```



## <a name="Worker">type</a> [Worker](https://github.com/cognusion/go-jar/tree/master/workers/worker.go?s=258:545#L11)
``` go
type Worker struct {
    // WorkChan is where the work comes from
    WorkChan chan Work
    // QuitChan will get some bools sent to it when the Worker pool needs to shrink
    QuitChan chan bool
    // KillChan will close when all the Workers need to exit
    KillChan chan struct{}
    // contains filtered or unexported fields
}

```
Worker is a simple primitive construct that listens on WorkChan for Work to do,
Might hear a "true" on QuitChan if it is underworked,
Might see a closed KillChan if it's time to leave expeditiously










### <a name="Worker.Do">func</a> (\*Worker) [Do](https://github.com/cognusion/go-jar/tree/master/workers/worker.go?s=615:636#L23)
``` go
func (w *Worker) Do()
```
Do forks off a Workerloop that listens for Work, quits, or kills




## <a name="WorkerPool">type</a> [WorkerPool](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=725:1399#L24)
``` go
type WorkerPool struct {
    // WorkChan is where the work goes
    WorkChan chan Work
    // Stop permanently stops the pool after current work is done.
    // WorkChan is not closed, to prevent errant reads
    Stop func()

    Metrics metrics.Meter
    // contains filtered or unexported fields
}

```
WorkerPool is an overly-complicated mechanation to allow arbitrary work to be accomplished by an arbitrary worker,
which will then return arbitrary results onto an arbitrary channel, while allowing for the evidence-driven growing or
shrinking of the pool of available workers based on the fillyness of the WorkChan, which should be buffered and of
an appropriate size. If that hasn't turned you off yet, carry on.







### <a name="NewWorkerPool">func</a> [NewWorkerPool](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=1875:1976#L52)
``` go
func NewWorkerPool(WorkChan chan Work, initialSize int, autoAdjustInterval time.Duration) *WorkerPool
```
NewWorkerPool returns a functioning WorkerPool bound to WorkChan, with an initial pool size of initialSize, and if autoAdjustInterval > 0, then
it will run the CheckAndAdjust() every that often.
NOTE: If your WorkChan is unbuffered (no size during make(), autoAdjust will not run, nor will calling CheckAndAdjust() result in changes. The channel capacity
and usage is key to this. It is recommended that the buffer size be around anticipated burst size for work





### <a name="WorkerPool.AddWorkers">func</a> (\*WorkerPool) [AddWorkers](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=6111:6156#L219)
``` go
func (p *WorkerPool) AddWorkers(number int64)
```
AddWorkers adds the specified number of workers




### <a name="WorkerPool.CheckAndAdjust">func</a> (\*WorkerPool) [CheckAndAdjust](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=3638:3675#L131)
``` go
func (p *WorkerPool) CheckAndAdjust()
```
CheckAndAdjust asynchronously triggers the process to possibly resize the pool.
While a resize process is running, subsequent processors will silently exit




### <a name="WorkerPool.Max">func</a> (\*WorkerPool) [Max](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=3394:3427#L125)
``` go
func (p *WorkerPool) Max(max int)
```
Max sets the maximum number of workers




### <a name="WorkerPool.Min">func</a> (\*WorkerPool) [Min](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=3270:3303#L120)
``` go
func (p *WorkerPool) Min(min int)
```
Min sets the minimum number of workers




### <a name="WorkerPool.RemoveWorkers">func</a> (\*WorkerPool) [RemoveWorkers](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=6471:6519#L233)
``` go
func (p *WorkerPool) RemoveWorkers(number int64)
```
RemoveWorkers removes the specified number of workers, or the number running.




### <a name="WorkerPool.Size">func</a> (\*WorkerPool) [Size](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=6918:6951#L249)
``` go
func (p *WorkerPool) Size() int64
```
Size returns the eventually-consistent number of workers in the pool




### <a name="WorkerPool.Work">func</a> (\*WorkerPool) [Work](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=3167:3198#L115)
``` go
func (p *WorkerPool) Work() int
```
Work returns the quantity of Work in the work channel








- - -
Generated by [godoc2md](http://godoc.org/github.com/cognusion/godoc2md)
