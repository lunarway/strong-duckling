package tcpchecker

import "time"

type LogReporter struct {
	lastReport time.Time
	lastOpen   bool
	Log        interface {
		Infof(msg string, args ...interface{})
		Debugf(msg string, args ...interface{})
	}
}

func (r *LogReporter) ReportPortCheck(report Report) {
	switch {
	case report.Open && r.lastOpen:
		// Port is still open - great
	case !report.Open && (r.lastOpen || r.lastReport == time.Time{}):
		// Port closed
		r.lastReport = time.Now()
		r.lastOpen = report.Open
		r.Log.Infof("Port %s:%v status is: %s", report.Address, report.Port, "closed")
	case report.Open && !r.lastOpen:
		// Port opened
		r.lastReport = time.Now()
		r.lastOpen = report.Open
		r.Log.Infof("Port %s:%v status is %s. Content in response: %s", report.Address, report.Port, "open", report.Content)
	case !report.Open && !r.lastOpen:
		// Port still closed
		if time.Now().Sub(r.lastReport) > 5*time.Minute {
			r.lastReport = time.Now()
			r.lastOpen = report.Open
			r.Log.Infof("Port %s:%v is still %v", report.Address, report.Port, "closed")
		}
	default:
		panic("This should never happen in LogReporter.ReportPortCheck")
	}
}
