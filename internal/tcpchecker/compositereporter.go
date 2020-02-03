package tcpchecker

type compositeReporter struct {
	reporters []Reporter
}

func (r compositeReporter) ReportPortCheck(report Report) {
	for _, reporter := range r.reporters {
		reporter.ReportPortCheck(report)
	}
}

func CompositeReporter(reporters ...Reporter) Reporter {
	return &compositeReporter{
		reporters: reporters,
	}
}
