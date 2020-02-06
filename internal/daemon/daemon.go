package daemon

import (
	"time"
)

// Configuration is a configuration struct specifying what and how a Daemon
// instance must run.
//
// Default values are set for all fields so they can be omitted. Be sure to set
// the Tick function though as without it nothing will ever be triggered by the
// daemon.
type Configuration struct {
	Probes   *Probes
	Interval time.Duration

	// Tick is the function called in every interval by the daemon.
	Tick func()
}

// Probes represents the available life cycle probes of a Daemon.
type Probes struct {
	Started func(time.Duration)
	Stopped func()
	Ticked  func()
	Skipped func()
}

func (c *Configuration) setDefaults() {
	if c.Probes == nil {
		c.Probes = &Probes{}
	}
	if c.Probes.Started == nil {
		c.Probes.Started = func(time.Duration) {}
	}
	if c.Probes.Stopped == nil {
		c.Probes.Stopped = func() {}
	}
	if c.Probes.Skipped == nil {
		c.Probes.Skipped = func() {}
	}
	if c.Probes.Ticked == nil {
		c.Probes.Ticked = func() {}
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
		d.config.Probes.Skipped()
	}
}

// Loop starts the daemon tick loop. It will run until provided stop
// channel is closed.
func (d *Daemon) Loop(stop chan struct{}) {
	d.config.Probes.Started(d.config.Interval)
	timer := time.NewTimer(d.config.Interval)
	d.askForTick()

	for {
		select {
		case <-stop:
			d.config.Probes.Stopped()
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
			d.config.Probes.Ticked()
			timer.Reset(d.config.Interval)
		case <-timer.C:
			// request a new tick in the tick buffer. This might be a noop if a tick
			// is already running.
			d.askForTick()
		}
	}
}
