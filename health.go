package jar

import (
	"github.com/cognusion/go-health"

	"github.com/rcrowley/go-metrics"

	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

const (
	// ErrNoSuchEntryError is returned by the Status Registry when no status exists for the requested thing
	ErrNoSuchEntryError = Error("no such element exists")

	// ErrNoSuchHealthCheckStatus is returned when a string-based status has been used, but no corresponding HealthCheckStatus exists
	ErrNoSuchHealthCheckStatus = Error("no such HealthCheckStatus exists")
)

var (
	// Metrics is a Registry for metrics, to be reported in the healthcheck
	Metrics = metrics.NewRegistry()
	// Status is a Registry for statuses, to be reported in the healthcheck
	Status = health.NewStatusRegistry()

	// NUMCPU is the number of CPUs at starttime
	NUMCPU = runtime.NumCPU()
	// GOVERSION is the version of Go
	GOVERSION = runtime.Version()

	// Counter is the clicker to a request counter.
	Counter func()

	// ThisProcess is updated information about this process
	ThisProcess *ProcessInfo

	// ConnectionCounter is used for tracking the current number of connections served
	ConnectionCounter int64

	// CurrentHealthCheck is a cache of the current state, refreshed periodically
	CurrentHealthCheck atomic.Value

	// HealthCheck is a Finisher that writes the healthcheck
	HealthCheck = healthCheckAsync

	// TerseHealthCheck is a Finisher that writes the terse healthcheck
	TerseHealthCheck = terseHealthCheckAsync
)

func init() {

	rc := metrics.NewRegisteredMeter("Requests", Metrics)
	Counter = func() { rc.Mark(1) }

	// Store an empty healthcheck
	CurrentHealthCheck.Store(health.NewCheck())

	InitFuncs.Add(func() {
		if !TaskRegistry.Exists("HealthCheck Cacher") {
			// Since InitFuncs may be called multiple times, we don't want to orphan these
			TaskRegistry.AddEvery("HealthCheck Cacher", func() error {
				// Build a healthcheck
				hc := health.NewCheck()
				CurrentHealthCheck.Store(*getHC(&hc))

				if Conf.GetBool(ConfigDebug) {
					var (
						cpu  float64
						mem  float64
						size int64
						rate float64
					)

					if Workers != nil {
						size = Workers.Size()
						rate = Workers.Metrics.Rate1()
					}
					if ThisProcess != nil {
						cpu = ThisProcess.CPU()
						mem = ThisProcess.Memory()
					}
					DebugOut.Printf("Rate: %.4f/second Goros: %d Workers: %d Worker Rate: %.2f Connections: %d Requests: %d CPU: %.2f%% RAM: %.2f%%\n", rc.Rate1(), runtime.NumGoroutine(), size, rate, ConnectionCounterGet(), rc.Count(), cpu, mem)
				}
				return nil
			}, 10*time.Second)
		}

		if ThisProcess == nil {
			// Since InitFuncs may be called multiple times, we don't want to orphan these
			ThisProcess = NewProcessInfo(0)
			ctx, cf := context.WithCancel(context.Background())
			ThisProcess.Ctx = ctx
			StopFuncs.Add(func() { // Add the context canceller
				DebugOut.Printf("Stopping ProcessInfo...\n")
				cf()
			})
			go func() {
				defer Status.Add("ProcessInfo", "WARNING", "ProcessInfo UpdateCpu stopped", nil)
				ThisProcess.UpdateCPU()
			}()
		}
	})

	// Link our Finishers
	Finishers["healthcheck"] = HealthCheck
	Finishers["tersehealthcheck"] = TerseHealthCheck
	Finishers["stack"] = Stack

}

// HealthCheckStatus is a specific int for HealthCheckStatus consts
type HealthCheckStatus int

func (i HealthCheckStatus) String() string {
	if k, ok := _HealthCheckStatusMap[i]; ok {
		return k
	}
	return fmt.Sprintf("HealthCheckStatus(%d)", i)
}

// Constants for HealthCheckStatuses
const (
	Unknown HealthCheckStatus = iota
	Ok
	Warning
	Critical
)

var _HealthCheckStatusMap = map[HealthCheckStatus]string{
	Unknown:  "Unknown",
	Ok:       "Ok",
	Warning:  "Warning",
	Critical: "Critical",
}

// StringToHealthCheckStatus takes a string HealthCheckStatus and returns the HealthCheckStatus or ErrNoSuchHealthCheckStatus
func StringToHealthCheckStatus(hc string) (HealthCheckStatus, error) {
	lchc := strings.ToLower(hc)
	for i, s := range _HealthCheckStatusMap {
		if lchc == strings.ToLower(s) {
			return i, nil
		}
	}
	return 0, ErrNoSuchHealthCheckStatus
}

// healthCheckAsync is a Finisher that writes out a cached healthcheck
func healthCheckAsync(w http.ResponseWriter, r *http.Request) {

	// Set the header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Grab the Check
	hc := CurrentHealthCheck.Load().(health.Check)

	// Send it, as JSON
	fmt.Fprint(w, hc.JSON())

}

// terseHealthCheckAsync is a Finisher that writes out a cached healthcheck
func terseHealthCheckAsync(w http.ResponseWriter, r *http.Request) {

	// Set the header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Grab the Check
	hc := CurrentHealthCheck.Load().(health.Check)

	// Send it, as JSON
	fmt.Fprint(w, hc.Terse())

}

// healthCheckSync is a Finisher that assembles healthcheck data, formats it, and writes it out
func healthCheckSync(w http.ResponseWriter, r *http.Request) {

	// Set the header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Grab the Check
	hc := health.NewCheck()
	hc = *getHC(&hc)

	// Send it, as JSON
	fmt.Fprint(w, hc.JSON())

}

// Stack is a Finisher that dumps the current stack to the request
func Stack(w http.ResponseWriter, r *http.Request) {

	buf := make([]byte, 1<<16)
	runtime.Stack(buf, true)

	w.Write(buf)

}

// ConnectionCounterAdd atomically adds 1 to the ConnectionCounter
func ConnectionCounterAdd() {
	atomic.AddInt64(&ConnectionCounter, 1)
}

// ConnectionCounterRemove atomically adds -1 to the ConnectionCounter
func ConnectionCounterRemove() {
	atomic.AddInt64(&ConnectionCounter, -1)
}

// ConnectionCounterGet atomically returns the current value of the ConnectionCounter
func ConnectionCounterGet() int64 {
	return atomic.LoadInt64(&ConnectionCounter)
}

// getProcessCPU returns the current process usage for the interval, or an error.
// The returned value may be 0, briefly, before the first tracking interval has
// occurred.
func getProcessCPU() (float64, error) {
	return ThisProcess.CPU(), nil
}

// getProcessMemory returns the current process memory usage as a percentage of total, or an error.
func getProcessMemory() (float64, error) {
	return ThisProcess.Memory(), nil
}

// getHC takes an Check, adds a bunch of stuff to it, and returns it
func getHC(hc *health.Check) *health.Check {

	if hc == nil {
		nhc := health.NewCheck()
		hc = &nhc
	}
	if hc.OverallStatus == "UNKNOWN" {
		hc.OverallStatus = "OK"
	}

	hc.AddSystem(&health.Status{
		Name:  MacroDictionary.Replacer("%%NAME Version"),
		Value: MacroDictionary.Replacer("%%VERSION"),
	})
	hc.AddSystem(&health.Status{
		Name:  "Go Version",
		Value: GOVERSION,
	})

	hc.AddSystem(&health.Status{
		Name:  "CPU Count",
		Value: NUMCPU,
	})

	hc.AddMetric(&health.Status{
		Name:   "Goros",
		Status: "OK",
		Value:  runtime.NumGoroutine(),
	})

	hc.AddMetric(&health.Status{
		Name:   "Connections",
		Status: "OK",
		Value:  ConnectionCounterGet(),
	})

	if cpuu, err := getProcessCPU(); err == nil {
		hc.AddMetric(&health.Status{
			Name:   "CPU Usage",
			Status: "OK",
			Value:  fmt.Sprintf("%f%%", cpuu),
		})
	} else {
		hc.AddMetric(&health.Status{
			Name:   "CPU Usage",
			Status: "WARNING",
			Value:  err,
		})
	}

	if memu, err := getProcessMemory(); err == nil {
		hc.AddMetric(&health.Status{
			Name:   "Memory Usage",
			Status: "OK",
			Value:  fmt.Sprintf("%f%%", memu),
		})
	} else {
		hc.AddMetric(&health.Status{
			Name:   "Memory Usage",
			Status: "WARNING",
			Value:  err,
		})
	}

	// Add all the metrics from the MetricsRegistry
	hc = AddMetrics(Metrics.GetAll(), hc)

	// Add all the statuses from the StatusRegistry
	hc = AddStatuses(Status, hc)

	// Calculate new OverallStatus
	hc.Calculate()
	return hc
}

// AddStatuses ranges over the supplied StatusRegistry, adding each as a Service to the supplied Check
func AddStatuses(s *health.StatusRegistry, hc *health.Check) *health.Check {
	for _, key := range s.Keys() {
		status, err := s.Get(key)
		if err == nil {
			hc.AddService(status)
		}
	}
	return hc
}

// AddMetrics ranges over the supplied map, adding each as a Metric to the supplied Check
func AddMetrics(m map[string]map[string]interface{}, hc *health.Check) *health.Check {
	for name, vmap := range m {
		// map[count:1 1m.rate:0 5m.rate:0 15m.rate:0 mean.rate:0.35600778189293464]
		for metric, v := range vmap {
			smetric := health.SafeLabel(metric)
			mname := fmt.Sprintf("%s_%s", name, smetric)
			lsmetric := strings.ToLower(smetric)

			if strings.HasSuffix(lsmetric, "count") || strings.HasSuffix(lsmetric, "counter") {
				v = fmt.Sprintf("%d%s", v, "c")
			} else if name == "RequestTimes" && (!strings.HasSuffix(lsmetric, "rate") && !strings.HasSuffix(lsmetric, "stddev")) {
				switch vc := v.(type) {
				case float64:
					v = fmt.Sprintf("%d%s", int64(vc/1000000), "ms")
				case float32:
					v = fmt.Sprintf("%d%s", int64(vc/1000000), "ms")
				case int:
					v = fmt.Sprintf("%d%s", int64(vc/1000000), "ms")
				case int32:
					v = fmt.Sprintf("%d%s", int64(vc/1000000), "ms")
				case int64:
					v = fmt.Sprintf("%d%s", int64(vc/1000000), "ms")
				}
			} else if name == "RequestTimes" && (strings.HasSuffix(lsmetric, "rate") || strings.HasSuffix(lsmetric, "stddev")) {
				// We don't need these from RequestTimes. They're superfluous and slightly less accurate due to
				// where they are accrued in the request process
				continue
			}
			hc.AddMetric(&health.Status{
				Name:  mname,
				Value: v,
			})
		}
	}
	return hc
}
