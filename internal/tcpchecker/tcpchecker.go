package tcpchecker

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

func Check(name string, address string, port int, reporter Reporter) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%v", address, port), 1*time.Second)
	if err != nil {
		reporter.ReportPortCheck(Report{
			Name:    name,
			Address: address,
			Port:    port,
			Open:    false,
			Status:  "Connect error",
			Error:   err,
			Content: "",
		})
		return
	}

	err = conn.SetDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		reporter.ReportPortCheck(Report{
			Name:    name,
			Address: address,
			Port:    port,
			Open:    false,
			Status:  "Set deadline error",
			Error:   err,
			Content: "",
		})
		return
	}

	scanner := bufio.NewScanner(conn)

	output := strings.Builder{}
	for scanner.Scan() {
		fmt.Fprintf(&output, "%s\n", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		var netError net.Error
		if errors.As(err, &netError) && netError.Timeout() {
			reporter.ReportPortCheck(Report{
				Name:    name,
				Address: address,
				Port:    port,
				Open:    true,
				Status:  "Open (closed by us)",
				Error:   nil,
				Content: output.String(),
			})
			return
		}
		reporter.ReportPortCheck(Report{
			Name:    name,
			Address: address,
			Port:    port,
			Open:    false,
			Status:  "Scanner error",
			Error:   err,
			Content: output.String(),
		})
		return
	}
	reporter.ReportPortCheck(Report{
		Name:    name,
		Address: address,
		Port:    port,
		Open:    true,
		Status:  "Open (closed by peer)",
		Error:   nil,
		Content: output.String(),
	})
}
