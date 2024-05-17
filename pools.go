package jar

import (
	"github.com/cognusion/go-jar/workers"

	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	// ErrNoSuchMemberError is returned if the member doesn't exist or has been removed from a Pool
	ErrNoSuchMemberError = Error("member no longer exists in pool")
	// NoPoolsError is returned if there are no pools, but a Build was requested
	NoPoolsError = Error("there are no pools to build")
)

// Constants for configuration key strings
const (
	ConfigPoolsDefaultMemberErrorStatus               = ConfigKey("pools.defaultmembererrorstatus")
	ConfigPoolsDefaultMemberWeight                    = ConfigKey("pools.defaultmemberweight")
	ConfigPoolsHealthcheckInterval                    = ConfigKey("pools.healthcheckinterval")
	ConfigPoolsLocalMemberWeight                      = ConfigKey("pools.localmemberweight")
	ConfigPoolsPreMaterialize                         = ConfigKey("pools.prematerialize")
	ConfigPoolsDefaultConsistentHashPartitions        = ConfigKey("pools.defaultconsistenthashpartitions")
	ConfigPoolsDefaultConsistentHashReplicationFactor = ConfigKey("pools.defaultconsistenthashreplicationfactor")
	ConfigPoolsDefaultConsistentHashLoad              = ConfigKey("pools.defaultconsistenthashload")
)

func init() {
	InitFuncs.Add(func() {
		workers.DebugOut = DebugOut
	})

	ConfigAdditions[ConfigPoolsHealthcheckInterval] = 1 * time.Minute
	ConfigAdditions[ConfigPoolsPreMaterialize] = false
}

// Pools is a goro-safe map of Pool objects, and if interval > 0, will also
// healthcheck pool members, managing them accordingly.
type Pools struct {
	sync.RWMutex                   // Readers must RLock/RUnlock. Writers must Lock/Unlock
	pools         map[string]*Pool // Where the Pools are
	checkInterval time.Duration    // How often to check pool members
	// StopWatch will stop the monitoring of the pool members.
	StopWatch func()
	stopChan  chan struct{}
}

// NewPools creates a functioning Pools struct, initialized with the pools, and a healthcheck interval.
// Set the interval to 0 to disable healthchecks
func NewPools(poolConfigs map[string]*PoolConfig, interval time.Duration) (*Pools, error) {

	pools := poolConfigMapToPoolMap(poolConfigs)

	p := Pools{
		pools:         pools,
		checkInterval: interval,
		stopChan:      make(chan struct{}),
	}

	if interval > 0 && len(pools) > 0 {
		p.StopWatch = func() {
			p.StopWatch = func() {}
			DebugOut.Printf("Stopping Pool Lifeguard\n")
			close(p.stopChan)
		}
		StopFuncs.Add(p.StopWatch)
		go func() {
			defer Status.Add("Pool Lifeguard", "WARNING", "Lifeguard asked to leave the Pool", nil)
			p.healthTicker()
		}()
	}

	if Conf.GetBool(ConfigPoolsPreMaterialize) {
		for _, pool := range pools {
			_, err := pool.GetPool() // discard handler
			if err != nil {
				return nil, err
			}
		}
	}

	return &p, nil
}

// healthTicker fires up a ticker to deal with healthchecks. Never call this twice unless you know what you're doing.
func (p *Pools) healthTicker() {
	if p.checkInterval <= 0 {
		// Safety
		return
	}

	if len(p.pools) <= 0 {
		// safety
		return
	}

	// TODO: Size this appropriately
	var (
		rChan    = make(chan interface{}, 100)
		hcErrors = make(map[string]*HealthCheckError)
	)

	// Get the ticker going
	ticker := time.NewTicker(p.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			// Kill signalled
			return
		case r := <-rChan:
			// TODO: Flap detection
			// something interesting has arrived
			DebugOut.Printf("HC Returned: %T: '%v'\n", r, r)
			switch t := r.(type) {
			case HealthCheckError:
				herr := r.(HealthCheckError)
				if _, ok := hcErrors[fmt.Sprintf("%s %s", herr.PoolName, herr.URL)]; !ok {
					// We don't have an error already for this pool, remove the member
					if herr.Prune {
						DebugOut.Printf("Pruning %s: Removing %s (%s)\n", herr.PoolName, herr.URL, herr.Error())
						herr.Remove(herr.URL)
					}
				}
				hcErrors[fmt.Sprintf("%s %s", herr.PoolName, herr.URL)] = &herr
				Status.Add(fmt.Sprintf("%s_%s", herr.PoolName, herr.URL), strings.ToUpper(herr.ErrorStatus.String()), herr.Error(), nil)
			case HealthCheckResult:
				hres := r.(HealthCheckResult)
				if _, ok := hcErrors[fmt.Sprintf("%s %s", hres.PoolName, hres.URL)]; ok {
					// We had an error here, but now don't
					if hres.Prune {
						DebugOut.Printf("Pruning %s: Adding %s\n", hres.PoolName, hres.URL)
						hres.Add(hres.URL)
					}
					delete(hcErrors, fmt.Sprintf("%s %s", hres.PoolName, hres.URL))
				}
				Status.Add(fmt.Sprintf("%s_%s", hres.PoolName, hres.URL), "OK", nil, nil)
			default:
				// Not possible?
				ErrorOut.Printf("HealthCheck returned impossible type %s : %+v\n", t, r)
			}
		case <-ticker.C:
			go p.tickFunc(rChan)
		}
	}
}

// List returns list of Pool names
func (p *Pools) List() []string {
	p.RLock()
	defer p.RUnlock()

	pl := make([]string, len(p.pools))
	c := 0
	for name := range p.pools {
		pl[c] = name
		c++
	}
	return pl
}

// Exists returns bool if the named Pool exists
func (p *Pools) Exists(name string) bool {
	p.RLock()
	_, ok := p.pools[name]
	p.RUnlock()

	return ok
}

// Get returns a Pool and a bool, given a name, from the Pools
func (p *Pools) Get(name string) (*Pool, bool) {
	p.RLock()
	defer p.RUnlock()

	if pool, ok := p.pools[name]; ok {
		return pool, true
	}
	return nil, false
}

// Set adds-or-replaces the named pool
func (p *Pools) Set(name string, pool *Pool) {
	p.Lock()
	defer p.Unlock()

	p.pools[name] = pool
}

// Merge adds-or-replaces the specified pools
func (p *Pools) Merge(pools map[string]*Pool) {
	p.Lock()
	defer p.Unlock()

	for name, pool := range pools {
		p.pools[name] = pool
	}
}

// Replace does exactly that on the entire map of Pool
func (p *Pools) Replace(pools map[string]*Pool) {
	p.Lock()
	defer p.Unlock()

	p.pools = pools
}

// tickFunc runs every tick of the healthcheck
func (p *Pools) tickFunc(rChan chan interface{}) {
	DebugOut.Printf("Pools.healthTicker firing...\n")
	var (
		stagger  time.Duration
		worklist []*HealthCheckWork
	)

	// Quickly traverse the pools to add work to our list
	p.RLock()
	for _, pool := range p.pools {

		f := func(u, m interface{}) bool {
			murl := u.(url.URL)
			//member := m.(*Member)

			if !pool.Config.HealthCheckDisabled && pool.Config.HealthCheckURI != "" {
				hcurl := fmt.Sprintf("%s://%s%s", murl.Scheme, murl.Host, pool.Config.HealthCheckURI)
				if pool.Config.HealthCheckShotgun {
					// Don't schedule it, just fire it off now
					DebugOut.Printf("\tAdding immediate work for '%s'\n", hcurl)
					AddWork(&HealthCheckWork{
						PoolName:    pool.Config.Name,
						Member:      murl.String(),
						URL:         hcurl,
						ReturnChan:  rChan,
						Prune:       pool.Config.Prune,
						ErrorStatus: pool.healthCheckErrorStatus,
						Add:         pool.AddMember,
						Remove:      pool.RemoveMember,
					})
				} else {
					// Schedule it
					DebugOut.Printf("\tAdding scheduled work for Pool %s : %s\n", pool.Config.Name, murl.String())
					worklist = append(worklist, &HealthCheckWork{
						PoolName:    pool.Config.Name,
						Member:      murl.String(),
						URL:         hcurl,
						ReturnChan:  rChan,
						Prune:       pool.Config.Prune,
						ErrorStatus: pool.healthCheckErrorStatus,
						Add:         pool.AddMember,
						Remove:      pool.RemoveMember,
					})
				}
			}
			return true
		}

		// if the Pool is Materialized, the healthcheck is enabled, and the URI isn't empty...
		if pool.IsMaterialized() && !pool.Config.HealthCheckDisabled && pool.Config.HealthCheckURI != "" {
			if len(pool.ListMembers()) == 0 {
				// Never ever ever have an empty pool
				Status.Add(pool.Config.Name, "CRITICAL", "Pool has no members", nil)
			} else if _, err := Status.Get(pool.Config.Name); err == nil {
				// We had an empty pool, but it's all over now
				Status.Remove(pool.Config.Name)
			}
		}

		// Iterate over the members
		pool.members.Range(f)
	}
	p.RUnlock()

	if len(worklist) > 0 {
		// Calculate how long we should wait between things to do
		stagger = time.Second * time.Duration(p.checkInterval.Seconds()/float64(len(worklist)))
		DebugOut.Printf("Pool Member Healthcheck Stagger is %s\n", stagger.String())

		// Iterate through the to-monitor members, and add the work,
		// delaying based on the calculated stagger
		var c int64 = 1
		for _, w := range worklist {

			go func(w *HealthCheckWork, delay time.Duration) {
				DebugOut.Printf("Adding work for '%s' in %s\n", w.URL, delay.String())
				<-time.After(delay)
				AddWork(w)
			}(w, time.Duration(c)*stagger)

			c++
		}
	}
}

func poolConfigMapToPoolMap(poolConfigs map[string]*PoolConfig) map[string]*Pool {
	p := make(map[string]*Pool)
	for name, config := range poolConfigs {
		p[name] = NewPool(config)
	}
	return p
}

// BuildPools unmarshalls the pools config, creates them, and updates the pool list
// ConfigPoolsHealthcheckInterval will set the healthcheck interval for pool members.
// Set to 0 to disable.
func BuildPools() (*Pools, error) {

	if ipools := Conf.Get(ConfigPools); ipools != nil {
		// We have pools in the config
		pools := make(map[string]*PoolConfig)
		Conf.UnmarshalKey(ConfigPools, &pools)
		DebugOut.Printf("Pools %+v\n", pools)
		hcDuration := Conf.GetDuration(ConfigPoolsHealthcheckInterval)
		return NewPools(pools, hcDuration)
	}

	return nil, NoPoolsError
}

// PruneFunc is a func that may add or remove Pool members
type PruneFunc func(string) error

// HealthCheckError is an error returned through the HealthCheck system
type HealthCheckError struct {
	PoolName    string
	URL         string
	StatusCode  int
	Prune       bool
	ErrorStatus HealthCheckStatus
	Add         PruneFunc
	Remove      PruneFunc
	Err         error
}

// Error returns the stringified version of the error
func (h *HealthCheckError) Error() string {
	return fmt.Sprintf("Code: %d Error: '%s'", h.StatusCode, h.Err)
}

// HealthCheckResult is a non-error returned through the HealthCheck system
type HealthCheckResult struct {
	PoolName   string
	URL        string
	StatusCode int
	Prune      bool
	Add        PruneFunc
	Remove     PruneFunc
}

// HealthCheckWork is Work to run a HealthCheck
type HealthCheckWork struct {
	PoolName    string
	Member      string
	URL         string
	Prune       bool
	ErrorStatus HealthCheckStatus
	Add         PruneFunc
	Remove      PruneFunc
	// Return is an error, or the StatusCode int
	ReturnChan chan interface{}
}

// Work executes the HealthCheck and returns HealthCheckResult or HealthCheckError
func (h *HealthCheckWork) Work() interface{} {
	var workClient = &http.Client{
		Timeout: time.Second * 2,
	}

	res, err := workClient.Get(h.URL)
	if err != nil {
		return HealthCheckError{
			PoolName:    h.PoolName,
			URL:         h.Member,
			Err:         err,
			Prune:       h.Prune,
			ErrorStatus: h.ErrorStatus,
			Add:         h.Add,
			Remove:      h.Remove,
		}
	}
	_, err = io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return HealthCheckError{
			PoolName:    h.PoolName,
			URL:         h.Member,
			StatusCode:  res.StatusCode,
			Err:         err,
			Prune:       h.Prune,
			ErrorStatus: h.ErrorStatus,
			Add:         h.Add,
			Remove:      h.Remove,
		}
	}

	if res.StatusCode > 299 {
		return HealthCheckError{
			PoolName:    h.PoolName,
			URL:         h.Member,
			StatusCode:  res.StatusCode,
			Err:         fmt.Errorf("%s", res.Status),
			Prune:       h.Prune,
			ErrorStatus: h.ErrorStatus,
			Add:         h.Add,
			Remove:      h.Remove,
		}
	}
	return HealthCheckResult{
		PoolName:   h.PoolName,
		URL:        h.Member,
		StatusCode: res.StatusCode,
		Prune:      h.Prune,
		Add:        h.Add,
		Remove:     h.Remove,
	}
}

// Return consumes a Work result and slides it downthe return channel
func (h *HealthCheckWork) Return(rthing interface{}) {
	h.ReturnChan <- rthing
}
