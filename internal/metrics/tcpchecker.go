package metrics

import (
	"github.com/lunarway/strong-duckling/internal/tcpchecker"
)

type TcpCheckerMetricsReporter struct {
}

func (r *TcpCheckerMetricsReporter) ReportPortCheck(report tcpchecker.Report) {
}
