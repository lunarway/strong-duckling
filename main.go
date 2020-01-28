package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/lunarway/strong-duckling/internal/daemon"
	"github.com/lunarway/strong-duckling/internal/http"
	"github.com/lunarway/strong-duckling/internal/metrics"
	"github.com/lunarway/strong-duckling/internal/whooping"
	"github.com/prometheus/common/log"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = ""
)

func main() {
	flags := kingpin.New("strong-duckling", "A small sidekick to strongswan VPN")
	listenAddress := flags.Flag("listen", "Address on which to expose metrics.").Default(":9100").String()
	whoopingAddress := flags.Flag("whooping", "Address on which to start whooping.").String()
	log.AddFlags(flags)
	flags.HelpFlag.Short('h')
	flags.Version(version)
	kingpin.MustParse(flags.Parse(os.Args[1:]))

	whooper := whooping.Whooper{}

	server := http.Define()
	whooper.RegisterListener(server, fmt.Sprintf("http://localhost%s", *listenAddress))
	metrics.Register(server)

	componentDone := make(chan error)
	shutdown := make(chan struct{})
	var shutdownWg sync.WaitGroup

	if whoopingAddress != nil && *whoopingAddress != "" {
		whoopDaemon := daemon.New(daemon.Configuration{
			Logger:   log.With("name", "whooper"),
			Interval: 1 * time.Second,
			Tick: func() {
				whooper.Whoop(*whoopingAddress, fmt.Sprintf("http://localhost%s", *listenAddress))
			},
		})

		shutdownWg.Add(1)
		go func() {
			defer shutdownWg.Done()
			whoopDaemon.Loop(shutdown)
		}()
	}

	go func() {
		// no shutdown mechanism in place for the HTTP server
		componentDone <- http.Start(server, *listenAddress)
	}()

	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-sigs:
			log.Infof("Received os signal '%s'. Terminating...", sig)
			componentDone <- nil
		case <-shutdown:
		}
	}()

	reason := <-componentDone
	if reason != nil {
		log.Errorf("exited due to error: %v", reason)
	} else {
		log.Info("exited due to a component shutting down")
	}
	close(shutdown)
	log.Info("waiting for all components to shutdown")
	shutdownWg.Wait()
	if reason != nil {
		os.Exit(1)
	}
}
