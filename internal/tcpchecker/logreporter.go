package tcpchecker

import (
	"time"

	"go.uber.org/zap"
)

type logReporter struct {
	lastReport time.Time
	lastOpen   bool
	Logger     zap.Logger
}

func (r *logReporter) ReportPortCheck(report Report) {
	l := r.Logger.Sugar().With("report", report)
	switch {
	case report.Open && r.lastOpen:
		// Port is still open - great
	case !report.Open && (r.lastOpen || r.lastReport == time.Time{}):
		// Port closed
		r.lastReport = time.Now()
		r.lastOpen = report.Open
		l.
			With("status", "closed").
			Infof("TCP connection to %s closed", report.Name)
	case report.Open && !r.lastOpen:
		// Port opened
		r.lastReport = time.Now()
		r.lastOpen = report.Open
		l.
			With("status", "opened").
			Infof("TCP connection to %s opened", report.Name)
	case !report.Open && !r.lastOpen:
		// Port still closed
		if time.Since(r.lastReport) > 5*time.Minute {
			r.lastReport = time.Now()
			r.lastOpen = report.Open
			l.
				With("status", "closed").
				Infof("TCP connection to %s is still closed", report.Name)
		}
	default:
		panic("This should never happen in LogReporter.ReportPortCheck")
	}
}

func LogReporter(logger zap.Logger) Reporter {
	return &logReporter{Logger: logger}
}
