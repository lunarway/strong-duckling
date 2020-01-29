package tcpchecker

import (
	"time"

	"github.com/lunarway/strong-duckling/internal/daemon"
	"github.com/prometheus/common/log"
)

type PortChecker struct {
	address string
	port    int
}

func StartPortChecking(address string, port int, reporter Reporter) *daemon.Daemon {
	logger := log.With("name", "portchecker").With("address", address).With("port", port)
	logger.Infof("Start checking address %s:%v", address, port)
	daemon := daemon.New(daemon.Configuration{
		Logger:   logger,
		Interval: 1 * time.Second,
		Tick: func() {
			str, err := CheckPort(address, int(port), reporter)
			logger.Debugf("Output: %s. Err %s", str, err)
		},
	})
	return daemon
}
