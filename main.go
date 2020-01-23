package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	done := make(chan error, 1)

	whooper := whooping.Whooper{}

	server := http.Define()
	whooper.RegisterListener(server, fmt.Sprintf("http://localhost%s", *listenAddress))
	metrics.Register(server)

	if whoopingAddress != nil && *whoopingAddress != "" {
		whooper.StartWhooping(*whoopingAddress, fmt.Sprintf("http://localhost%s", *listenAddress))
	}

	//defer http.Stop()
	go func() {
		done <- http.Start(server, *listenAddress)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Infof("Received os signal '%s'. Terminating...", sig)
		done <- nil
	}()

	reason := <-done
	if reason != nil {
		log.Errorf("exited due to error: %v", reason)
		os.Exit(1)
	}
	log.Debug("exited with exit 0")
}
