package tcpchecker

import (
	"time"

	"github.com/prometheus/common/log"
)

type logReporter struct {
	lastReport time.Time
	lastOpen   bool
	Logger     log.Logger
}

func (r *logReporter) ReportPortCheck(report Report) {
	switch {
	case report.Open && r.lastOpen:
		// Port is still open - great
	case !report.Open && (r.lastOpen || r.lastReport == time.Time{}):
		// Port closed
		r.lastReport = time.Now()
		r.lastOpen = report.Open
		r.Logger.
			With("status", "closed").
			Infof("TCP connection to %s closed", report.Name)
	case report.Open && !r.lastOpen:
		// Port opened
		r.lastReport = time.Now()
		r.lastOpen = report.Open
		r.Logger.
			With("status", "opened").
			With("content", report.Content).
			Infof("TCP connection to %s opened", report.Name)
	case !report.Open && !r.lastOpen:
		// Port still closed
		if time.Now().Sub(r.lastReport) > 5*time.Minute {
			r.lastReport = time.Now()
			r.lastOpen = report.Open
			r.Logger.
				With("status", "closed").
				Infof("TCP connection to %s is still closed", report.Name)
		}
	default:
		panic("This should never happen in LogReporter.ReportPortCheck")
	}
}

func LogReporter(logger log.Logger) Reporter {
	return &logReporter{Logger: logger}
}
