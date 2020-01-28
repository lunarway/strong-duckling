package daemon

import (
	"os"
	"time"

	"github.com/prometheus/common/log"
)

// Configuration is a configuration struct specifying what and how a Daemon
// instance must run.
//
// Default values are set for all fields so they can be omitted. Be sure to set
// the Tick function though as without it nothing will ever be triggered by the
// daemon.
type Configuration struct {
	Logger   log.Logger
	Interval time.Duration

	// Tick is the function called in every interval by the daemon.
	Tick func()
}

func (c *Configuration) setDefaults() {
	if c.Logger == nil {
		c.Logger = log.NewLogger(os.Stderr).With("name", "daemon")
	}
	if c.Interval == 0 {
		c.Interval = 5 * time.Minute
	}
	if c.Tick == nil {
		c.Tick = func() {}
	}
}

// Daemon provides a scheduled invokation of the configured Tick function. Start
// the daemon by call the blocking method Loop and stop it again by closing the
// provided stop channel.
type Daemon struct {
	config Configuration
	// tickSoon is a limited buffer of tick requests. Use method askForTick to
	// schedule new ticks through the buffer.
	tickSoon chan struct{}
}

// New allocates and returns an unstarted Daemon struct.
func New(c Configuration) *Daemon {
	c.setDefaults()
	d := Daemon{
		config:   c,
		tickSoon: make(chan struct{}, 1),
	}
	return &d
}

// askForTick requests a new tick. It ensures that only one tick can be running
// at any given time and drops tick requests if one is already running.
func (d *Daemon) askForTick() {
	select {
	case d.tickSoon <- struct{}{}:
	default:
		d.config.Logger.Info("Skipping tick as one is already in progress")
	}
}

// Loop starts the daemon tick loop. It will run until provided stop
// channel is closed.
func (d *Daemon) Loop(stop chan struct{}) {
	d.config.Logger.Info("Starting loop")
	timer := time.NewTimer(d.config.Interval)
	d.askForTick()

	for {
		select {
		case <-stop:
			d.config.Logger.Info("Stopping loop")
			// ensure to drain the timer channel before exiting as we don't know if
			// the shutdown is started before or after the timer have triggered.
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		case <-d.tickSoon:
			// ensure to drain the timer channel before ticking as we don't know if
			// the tick was scheduled by the timer or another event.
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			d.config.Tick()
			timer.Reset(d.config.Interval)
		case <-timer.C:
			d.config.Logger.Info("Tick timer asking for tick")
			// request a new tick in the tick buffer. This might be a noop if a tick
			// is already running.
			d.askForTick()
		}
	}
}
