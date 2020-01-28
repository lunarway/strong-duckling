package tcpchecker

type Reporter interface {
	ReportPortCheck(report Report)
}

type Report struct {
	Address string
	Port    int
	Open    bool
	Content string
	Status  string
	Error   error
}
