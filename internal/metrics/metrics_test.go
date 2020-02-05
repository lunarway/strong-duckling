package metrics

import "github.com/lunarway/strong-duckling/internal/tcpchecker"

var _ tcpchecker.Reporter = (&PrometheusReporter{}).TcpChecker()
