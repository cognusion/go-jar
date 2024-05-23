

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
  * [func (w *Worker) DoOnce(work Work)](#Worker.DoOnce)
* [type WorkerPool](#WorkerPool)
  * [func NewSimpleWorkerPool(WorkChan chan Work) *WorkerPool](#NewSimpleWorkerPool)
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










### <a name="Worker.Do">func</a> (\*Worker) [Do](https://github.com/cognusion/go-jar/tree/master/workers/worker.go?s=733:754#L28)
``` go
func (w *Worker) Do()
```
Do forks off a Workerloop that listens for Work, quits, or kills




### <a name="Worker.DoOnce">func</a> (\*Worker) [DoOnce](https://github.com/cognusion/go-jar/tree/master/workers/worker.go?s=609:643#L23)
``` go
func (w *Worker) DoOnce(work Work)
```
DoOnce asks the Worker to do one unit of Work and then end




## <a name="WorkerPool">type</a> [WorkerPool](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=1056:1751#L28)
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
WorkerPool was an overly-complicated mechanation to allow arbitrary work to be accomplished by an arbitrary worker.
If you obtain your WorkerPool using NewSimpleWorkerPool, most of the over-complicated mechanation is ignored.
If you obtain your WorkerPool using NewWorkerPool, the following is deprecated, and applies:

WorkerPool is an overly-complicated mechanation to allow arbitrary work to be accomplished by an arbitrary worker,
which will then return arbitrary results onto an arbitrary channel, while allowing for the evidence-driven growing or
shrinking of the pool of available workers based on the fillyness of the WorkChan, which should be buffered and of
an appropriate size. If that hasn't turned you off yet, carry on.







### <a name="NewSimpleWorkerPool">func</a> [NewSimpleWorkerPool](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=1967:2023#L55)
``` go
func NewSimpleWorkerPool(WorkChan chan Work) *WorkerPool
```
NewSimpleWorkerPool returns a functioning WorkerPool bound to WorkChan, that does not try keep Workers running. It is strongly recommended that
WorkChan be unbounded unless you demonstrably need to bound it.


### <a name="NewWorkerPool">func</a> [NewWorkerPool](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=2911:3012#L80)
``` go
func NewWorkerPool(WorkChan chan Work, initialSize int, autoAdjustInterval time.Duration) *WorkerPool
```
NewWorkerPool returns a functioning WorkerPool bound to WorkChan, with an initial pool size of initialSize, and if autoAdjustInterval > 0, then
it will run the CheckAndAdjust() every that often.
NOTE: If your WorkChan is unbuffered (no size during make(), autoAdjust will not run, nor will calling CheckAndAdjust() result in changes. The channel capacity
and usage is key to this. It is recommended that the buffer size be around anticipated burst size for work

Deprecated: NewWorkerPool has been replaced by NewSimpleWorkerPool. In v2 this version will be removed completely in lieu of the "simple" variety.





### <a name="WorkerPool.AddWorkers">func</a> (\*WorkerPool) [AddWorkers](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=7327:7372#L260)
``` go
func (p *WorkerPool) AddWorkers(number int64)
```
AddWorkers adds the specified number of workers

Deprecated: AddWorkers will be removed in v2




### <a name="WorkerPool.CheckAndAdjust">func</a> (\*WorkerPool) [CheckAndAdjust](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=4871:4908#L171)
``` go
func (p *WorkerPool) CheckAndAdjust()
```
CheckAndAdjust asynchronously triggers the process to possibly resize the pool.
While a resize process is running, subsequent processors will silently exit

Deprecated: CheckAndAdjust will be removed in v2




### <a name="WorkerPool.Max">func</a> (\*WorkerPool) [Max](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=4545:4578#L160)
``` go
func (p *WorkerPool) Max(max int)
```
Max sets the maximum number of workers

Deprecated: Max will be removed in v2




### <a name="WorkerPool.Min">func</a> (\*WorkerPool) [Min](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=4350:4383#L150)
``` go
func (p *WorkerPool) Min(min int)
```
Min sets the minimum number of workers

Deprecated: Min will be removed in v2




### <a name="WorkerPool.RemoveWorkers">func</a> (\*WorkerPool) [RemoveWorkers](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=7768:7816#L280)
``` go
func (p *WorkerPool) RemoveWorkers(number int64)
```
RemoveWorkers removes the specified number of workers, or the number running

Deprecated: RemoveWorkers will be removed in v2




### <a name="WorkerPool.Size">func</a> (\*WorkerPool) [Size](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=8317:8350#L302)
``` go
func (p *WorkerPool) Size() int64
```
Size returns the eventually-consistent number of workers in the pool. Returns -1 for simple pools

Deprecated: Size will be removed in v2




### <a name="WorkerPool.Work">func</a> (\*WorkerPool) [Work](https://github.com/cognusion/go-jar/tree/master/workers/workerpool.go?s=4203:4234#L143)
``` go
func (p *WorkerPool) Work() int
```
Work returns the quantity of Work in the work channel








- - -
Generated by [godoc2md](http://github.com/cognusion/godoc2md)
