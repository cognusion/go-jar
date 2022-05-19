package jar

import (
	"github.com/cognusion/go-cronzilla"

	"time"
)

var (
	// TaskRegistry is for wrangling scheduled tasks
	TaskRegistry cronzilla.Wrangler
)

func init() {
	StopFuncs.Add(TaskRegistry.Close)

	TaskRegistry.AddEvery("Wrangler Cleaner", func() error {
		TaskRegistry.Clean()
		return nil
	}, time.Minute)
}
