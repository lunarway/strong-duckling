package tcpchecker

type CompositeReporter struct {
	Reporters []Reporter
}

func (r CompositeReporter) ReportPortCheck(report Report) {
	for _, reporter := range r.Reporters {
		reporter.ReportPortCheck(report)
	}
}
