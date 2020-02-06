package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/lunarway/strong-duckling/internal/daemon"
	"github.com/lunarway/strong-duckling/internal/http"
	"github.com/lunarway/strong-duckling/internal/metrics"
	"github.com/lunarway/strong-duckling/internal/strongswan"
	"github.com/lunarway/strong-duckling/internal/tcpchecker"
	"github.com/lunarway/strong-duckling/internal/vici"
	"github.com/lunarway/strong-duckling/internal/whooping"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = ""
)

func main() {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()
	flags := kingpin.New("strong-duckling", "A small sidekick to strongswan VPN")
	listenAddress := flags.Flag("listen", "Address on which to expose metrics.").String()
	whoopingAddress := flags.Flag("whooping", "Address on which to start whooping.").String()
	tcpCheckerAddresses := flags.Flag("tcp-checker", "TCP address to check. Supports <address>:<port> or <name>:<address>:<port>").Strings()
	log.AddFlags(flags)
	flags.HelpFlag.Short('h')
	flags.Version(version)
	socket := flags.Flag("vici-socket", "VICI (charon.vici) socket to connect to. Usually /var/run/charon.vici").String()
	kingpin.MustParse(flags.Parse(os.Args[1:]))

	whooper := whooping.Whooper{}

	httpServer := http.Define()
	if *listenAddress != "" {
		whooper.RegisterListener(httpServer, fmt.Sprintf("http://localhost%s", *listenAddress))
		metrics.Register(httpServer)
	}
	prometheusReporter, err := metrics.NewPrometheusReporter(prometheus.DefaultRegisterer, log.Base().With("name", "prometheusReporter"))
	if err != nil {
		log.Errorf("Failed to register metrics: %v", err)
		os.Exit(1)
	}

	componentDone := make(chan error)
	shutdown := make(chan struct{})
	var shutdownWg sync.WaitGroup

	if whoopingAddress != nil && *whoopingAddress != "" {
		logger := log.With("name", "whooper")
		whoopDaemon := daemon.New(daemon.Configuration{
			Probes: &daemon.Probes{
				Started: func(d time.Duration) {
					logger.Infof("Whoop daemon started with interval %v", d)
					prometheusReporter.Daemon.Started.WithLabelValues("whooper", d.String()).Inc()
				},
				Stopped: func() {
					logger.Info("Whoop daemon stopped")
					prometheusReporter.Daemon.Stopped.WithLabelValues("whooper").Inc()
				},
				Skipped: func() {
					prometheusReporter.Daemon.Skipped.WithLabelValues("whooper").Inc()
				},
				Ticked: func() {
					prometheusReporter.Daemon.Ticked.WithLabelValues("whooper").Inc()
				},
			},
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

	for _, tcpCheckerAddress := range *tcpCheckerAddresses {
		values := strings.Split(tcpCheckerAddress, ":")
		var name, address, portStr string

		if len(values) == 3 {
			name = values[0]
			address = values[1]
			portStr = values[2]
		} else if len(values) == 2 {
			address = values[0]
			portStr = values[1]
			name = fmt.Sprintf("%s:%s", address, portStr)
		} else {
			log.Errorf("Could not understand tcp-checker %s", tcpCheckerAddress)
			os.Exit(1)
		}
		port, err := strconv.ParseInt(portStr, 10, 32)
		if err != nil {
			log.Errorf("Could not parse port %s as integer in tcp-checker %s. Error: %s", portStr, tcpCheckerAddress, err)
			os.Exit(1)
		}

		logger := log.
			With("type", "tcpchecker").
			With("name", name).
			With("address", address).
			With("port", port)
		logger.Infof("Start checking address %s:%v", address, port)
		tcpCheckerDaemon := daemon.New(daemon.Configuration{
			Probes: &daemon.Probes{
				Started: func(d time.Duration) {
					logger.Infof("tcpchecker daemon started with interval %v", d)
					prometheusReporter.Daemon.Started.WithLabelValues("tcpchecker", d.String()).Inc()
				},
				Stopped: func() {
					logger.Info("tcpchecker daemon stopped")
					prometheusReporter.Daemon.Stopped.WithLabelValues("tcpchecker").Inc()
				},
				Skipped: func() {
					prometheusReporter.Daemon.Skipped.WithLabelValues("tcpchecker").Inc()
				},
				Ticked: func() {
					prometheusReporter.Daemon.Ticked.WithLabelValues("tcpchecker").Inc()
				},
			},
			Interval: 1 * time.Second,
			Tick: func() {
				tcpchecker.Check(name, address, int(port), tcpchecker.CompositeReporter(tcpchecker.LogReporter(logger), prometheusReporter.TcpChecker()))
			},
		})

		shutdownWg.Add(1)
		go func() {
			defer shutdownWg.Done()
			tcpCheckerDaemon.Loop(shutdown)
			componentDone <- nil
		}()
	}

	if *listenAddress != "" {
		go func() {
			// no shutdown mechanism in place for the HTTP server
			componentDone <- http.Start(httpServer, *listenAddress)
		}()
	}

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

	if len(*socket) != 0 {
		conn, err := net.Dial("unix", *socket)
		if err != nil {
			log.Errorf("Failed to establish socket connection to vici: %v", err)
			os.Exit(1)
		}
		defer conn.Close()
		client := vici.NewClientConn(conn)
		defer client.Close()
		logger := log.With("name", "strongswan")
		d := daemon.New(daemon.Configuration{
			Probes: &daemon.Probes{
				Started: func(d time.Duration) {
					logger.Infof("strongswan daemon started with interval %v", d)
					prometheusReporter.Daemon.Started.WithLabelValues("strongswan", d.String()).Inc()
				},
				Stopped: func() {
					logger.Info("strongswan daemon stopped")
					prometheusReporter.Daemon.Stopped.WithLabelValues("strongswan").Inc()
				},
				Skipped: func() {
					prometheusReporter.Daemon.Skipped.WithLabelValues("strongswan").Inc()
				},
				Ticked: func() {
					prometheusReporter.Daemon.Ticked.WithLabelValues("strongswan").Inc()
				},
			},
			Interval: 2 * time.Second,
			Tick: func() {
				strongswan.Collect(client, prometheusReporter)
			},
		})

		go func() {
			d.Loop(shutdown)
			log.Infof("vici strongswan checker daemon stopped. Terminating...")
			componentDone <- nil
		}()
	}

	log.Infof("Strong duckling version %s", version)
	prometheusReporter.Info(version)

	// this is blocking until some component fails of a signal is received
	reason := <-componentDone

	close(shutdown)
	log.Info("waiting for all components to shutdown")
	shutdownWg.Wait()
	if reason != nil {
		log.Errorf("exited due to error: %v", reason)
		exitCode = 1
	} else {
		log.Info("exited due to a component shutting down")
	}
}
