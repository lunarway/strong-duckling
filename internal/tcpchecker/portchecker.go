package tcpchecker

import (
	"time"

	"github.com/prometheus/common/log"
)

type PortChecker struct {
	address string
	port    int
}

func StartPortChecking(address string, port int, reporter Reporter) PortChecker {

	log.Infof("Start checking address %s:%v", address, port)

	go func() {
		for true {
			// output, err :=
			CheckPort(address, int(port), reporter)
			// if err != nil {
			// 	log.Debugf("Failed connecting to address %s:%v. Error: %s", address, port, err)
			// } else {
			// 	log.Debugf("Successfully connected to address %s:%v. Content: %s", address, port, output)
			// }
			time.Sleep(1 * time.Second)
		}
	}()

	return PortChecker{}
}
