package jar

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/process"
	"go.uber.org/atomic"
)

// ProcessInfo is used to track information about ourselves.
// All member functions are safe to use across goros
type ProcessInfo struct {
	Ctx      context.Context
	cpu      *atomic.Float64
	mem      *atomic.Float64
	interval time.Duration
	pid      int32
	proc     *process.Process
}

var (
	// RestartSelf is a niladic that will trigger a graceful restart of this process
	RestartSelf func()
	// IntSelf is a niladic that will trigger an interrupt of this process
	IntSelf func()
	// KillSelf is a niladic that will trigger a graceful shutdown of this process
	KillSelf func()
)

func init() {
	Finishers["restart"] = Restart

	// Ensure we only restart ourselves once
	// why 10? small buffer so multiple requests close together hit the channel
	// and don't block the caller
	signalChan := make(chan os.Signal, 10)

	// These functions are used by handlers to trigger signals to the running
	// process.
	RestartSelf = func() { signalChan <- syscall.SIGUSR2 }
	IntSelf = func() { signalChan <- syscall.SIGINT }
	KillSelf = func() { signalChan <- syscall.SIGKILL }

	// We mirror the signal interception of grace, so we can properly clean up
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)

	go func() {
		for {
			select {
			case s := <-signalChan:
				DebugOut.Printf("%s signaled.", s.String())
				p, err := os.FindProcess(os.Getpid())
				if err != nil {
					ErrorOut.Printf("Error finding process '%d': %s\n", os.Getpid(), err)
				}
				p.Signal(s)
			case s := <-ch:
				switch s {
				case syscall.SIGINT, syscall.SIGTERM:
					StopFuncs.Call()
					signal.Stop(ch)
					return
				case syscall.SIGUSR2:
					StopFuncs.Call()
				}
			}
		}
	}()

}

// Restart signals the server to restart itself
func Restart(w http.ResponseWriter, r *http.Request) {
	RestartSelf()
	fmt.Fprint(w, "Done\n")
}

// NewProcessInfo returns an intialized ProcessInfo that has an interval set to 1 minute.
// Supply 0 as the pid to autodetect the running process' pid
func NewProcessInfo(pid int32) *ProcessInfo {
	if pid == 0 {
		pid = int32(os.Getpid())
	}
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil
	}
	return &ProcessInfo{
		proc:     p,
		interval: time.Minute,
		pid:      pid,
		cpu:      atomic.NewFloat64(0),
		mem:      atomic.NewFloat64(0),
	}

}

// SetInterval changes(?) the interval at which CPU slices are taken for comparison.
func (p *ProcessInfo) SetInterval(i time.Duration) {
	p.interval = i
}

// Memory returns the current value of the process memory, as a percent of total
func (p *ProcessInfo) Memory() float64 {
	return p.mem.Load()
}

// CPU returns the current value of the CPU tracker, as a percent of total
func (p *ProcessInfo) CPU() float64 {
	return p.cpu.Load()
}

// UpdateCPU loops while Ctx is valid, sampling our CPU usage every interval.
// This should generally only be called once, unless you know what you're doing
func (p *ProcessInfo) UpdateCPU() {
	for {
		select {
		case <-p.Ctx.Done():
			return
		default:
			if e, _ := process.PidExistsWithContext(p.Ctx, p.pid); !e {
				ErrorOut.Printf("Pid %d no longer exists. UpdateCPU exiting.\n", p.pid)
			}
			cpu, err := p.proc.PercentWithContext(p.Ctx, 0)
			if err != nil {
				ErrorOut.Printf("Error updating CPU usage: %s\n", err)
			} else if cpu == 0 {
				// init
			} else {
				p.cpu.Store(cpu)
			}

			m, err := p.proc.MemoryPercentWithContext(p.Ctx)
			if err != nil {
				ErrorOut.Printf("Error updating Memory usage: %s\n", err)
			} else {
				p.mem.Store(float64(m))
			}

			DebugOut.Printf("CPU %4f%% MEM %4f%%\n", cpu, m)
			time.Sleep(p.interval) //-start.Sub(time.Now())
		}
	}
}
