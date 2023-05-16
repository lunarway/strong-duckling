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

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/lunarway/strong-duckling/internal/daemon"
	"github.com/lunarway/strong-duckling/internal/http"
	"github.com/lunarway/strong-duckling/internal/metrics"
	"github.com/lunarway/strong-duckling/internal/strongswan"
	"github.com/lunarway/strong-duckling/internal/tcpchecker"
	"github.com/lunarway/strong-duckling/internal/vici"
	"github.com/lunarway/strong-duckling/internal/whooping"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	version = ""
)

func main() {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	flags := kingpin.New("strong-duckling", "A small sidekick to strongswan VPN")
	listenAddress := flags.Flag("listen", "Address on which to expose metrics.").String()
	whoopingAddress := flags.Flag("whooping", "Address on which to start whooping.").String()
	tcpCheckerAddresses := flags.Flag("tcp-checker", "TCP address to check. Supports <address>:<port> or <name>:<address>:<port>").Strings()
	enableReinitiator := flags.Flag("enable-reinitiator", "Enables re-initiation of connections when expected Security Associations are missing").Bool()
	flags.HelpFlag.Short('h')
	flags.Version(version)
	socket := flags.Flag("vici-socket", "VICI (charon.vici) socket to connect to. Usually /var/run/charon.vici").String()
	kingpin.MustParse(flags.Parse(os.Args[1:]))

	if *enableReinitiator && len(*socket) == 0 {
		zap.L().Sugar().Fatal("--enable-reinitiator requires --vici-socket to be set up")
		os.Exit(1)
	}

	whooper := whooping.Whooper{}

	httpServer := http.Define()
	if *listenAddress != "" {
		whooper.RegisterListener(httpServer, fmt.Sprintf("http://localhost%s", *listenAddress))
		metrics.Register(httpServer)
	}
	prometheusReporter, err := metrics.NewPrometheusReporter(prometheus.DefaultRegisterer, zap.L().With("name", "prometheusReporter"))
	if err != nil {
		zap.L().Sugar().Fatalf("Failed to register metrics: %v", err)
		os.Exit(1)
	}

	componentDone := make(chan error)
	shutdown := make(chan struct{})
	var shutdownWg sync.WaitGroup

	if whoopingAddress != nil && *whoopingAddress != "" {
		logger := zap.L().Sugar().With("name", "whooper")
		whoopDaemon := daemon.New(daemon.Configuration{
			Reporter: prometheusReporter.Daemon(logger, "whopper"),
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
			zap.L().Sugar().Errorf("Could not understand tcp-checker %s", tcpCheckerAddress)
			os.Exit(1)
		}
		port, err := strconv.ParseInt(portStr, 10, 32)
		if err != nil {
			zap.L().Sugar().Errorf("Could not parse port %s as integer in tcp-checker %s. Error: %s", portStr, tcpCheckerAddress, err)
			os.Exit(1)
		}

		logger := zap.L().Sugar().
			With("type", "tcpchecker").
			With("name", name).
			With("address", address).
			With("port", port)
		logger.Infof("Start checking address %s:%v", address, port)
		tcpCheckerDaemon := daemon.New(daemon.Configuration{
			Reporter: prometheusReporter.Daemon(logger, "tcpchecker"),
			Interval: 1 * time.Second,
			Tick: func() {
				tcpchecker.Check(name, address, int(port), tcpchecker.CompositeReporter(tcpchecker.LogReporter(logger), prometheusReporter.TcpChecker()))
			},
		})

		shutdownWg.Add(1)
		go func() {
			defer shutdownWg.Done()
			tcpCheckerDaemon.Loop(shutdown)
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
			zap.L().Sugar().Infof("Received os signal '%s'. Terminating...", sig)
			componentDone <- nil
		case <-shutdown:
		}
	}()

	if len(*socket) != 0 {
		ikeSAStatusReceivers := []strongswan.IKESAStatusReceiver{
			prometheusReporter.StrongSwan(),
		}

		if *enableReinitiator {
			reinitiatorClient := viciClient(&shutdownWg, shutdown, componentDone, zap.L().Sugar().With("viciClient", "reinitiator"), *socket)
			reinitiatorClient.ReadTimeout = 5 * time.Minute

			ikeSAStatusReceivers = append(ikeSAStatusReceivers, strongswan.NewReinitiator(reinitiatorClient, zap.L().Sugar().With("name", "reinitiator")))
		}

		client := viciClient(&shutdownWg, shutdown, componentDone, zap.L().Sugar().With("viciClient", "collector"), *socket)
		client.ReadTimeout = 60 * time.Second

		d := daemon.New(daemon.Configuration{
			Reporter: prometheusReporter.Daemon(zap.L().Sugar().With("name", "strongswan"), "strongswan"),
			Interval: 2 * time.Second,
			Tick: func() {
				strongswan.Collect(client, ikeSAStatusReceivers)
			},
		})

		shutdownWg.Add(1)
		go func() {
			defer shutdownWg.Done()
			d.Loop(shutdown)
			zap.L().Sugar().Infof("vici strongswan checker daemon stopped. Terminating...")
		}()
	}

	zap.L().Sugar().Infof("Strong duckling version %s", version)
	prometheusReporter.Info(version)

	// this is blocking until some component fails of a signal is received
	reason := <-componentDone

	close(shutdown)
	zap.L().Sugar().Info("waiting for all components to shutdown")
	shutdownWg.Wait()
	if reason != nil {
		zap.L().Sugar().Errorf("exited due to error: %v", reason)
		exitCode = 1
	} else {
		zap.L().Sugar().Info("exited due to a component shutting down")
	}
}

// viciClient returns a listening vici.ClientConn controlled by provided life
// cycle channels.
func viciClient(shutdownWg *sync.WaitGroup, shutdown chan struct{}, componentDone chan error, log zap.Logger, socket string) *vici.ClientConn {
	conn, err := net.Dial("unix", socket)
	if err != nil {
		zap.L().Sugar().Errorf("Failed to establish socket connection to vici on '%s': %v", socket, err)
		os.Exit(1)
	}
	client := vici.NewClientConn(conn)

	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		log.Info("vici client shutdown listener started")
		defer log.Info("vici client shutdown listener stopped")
		<-shutdown

		log.Info("Closing vici client listener")
		err := client.Close()
		if err != nil {
			log.Sugar().Errorf("Controlled close of vici client failed: %v", err)
		}
	}()

	shutdownWg.Add(1)
	go func() {
		defer shutdownWg.Done()
		log.Sugar().Infof("vici client listening on %s", socket)
		defer log.Info("vici client lister Go routine stopped")
		err := client.Listen()
		if err != nil {
			// we don't know if Listen stopped due to a controlled shutdown or due
			// to an underlying error. Log the error in the former case or report
			// the component done if the shutdown is unexpected
			select {
			case componentDone <- fmt.Errorf("vici client listener stopped unexpectedly: %w", err):
				return
			default:
				log.Sugar().Infof("vici client listener stopped: %v", err)
			}
		}
	}()

	return client
}
