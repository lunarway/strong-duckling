package tcpchecker

type Reporter interface {
	ReportPortCheck(report Report)
}

type Report struct {
	Name    string
	Address string
	Port    int
	Open    bool
	Content string
	Status  string
	Error   error
}
