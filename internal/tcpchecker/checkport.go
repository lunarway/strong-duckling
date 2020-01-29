package tcpchecker

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func CheckPort(address string, port int, reporter Reporter) (string, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%v", address, port), 1*time.Second)
	if err != nil {
		reporter.ReportPortCheck(Report{
			Address: address,
			Port:    port,
			Open:    false,
			Status:  "Connect error",
			Error:   err,
			Content: "",
		})
		return "", err
	}

	err = conn.SetDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		reporter.ReportPortCheck(Report{
			Address: address,
			Port:    port,
			Open:    false,
			Status:  "Set deadline error",
			Error:   err,
			Content: "",
		})
		return "", err
	}

	scanner := bufio.NewScanner(conn)

	output := ""
	for scanner.Scan() {
		output += fmt.Sprintf("%s\n", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		if netError, ok := err.(net.Error); ok && netError.Timeout() {
			reporter.ReportPortCheck(Report{
				Address: address,
				Port:    port,
				Open:    true,
				Status:  "Open",
				Error:   nil,
				Content: output,
			})
			return output, nil
		}
		reporter.ReportPortCheck(Report{
			Address: address,
			Port:    port,
			Open:    false,
			Status:  "Scanner error",
			Error:   err,
			Content: output,
		})
		return output, err
	}
	reporter.ReportPortCheck(Report{
		Address: address,
		Port:    port,
		Open:    true,
		Status:  "Open",
		Error:   nil,
		Content: output,
	})
	return output, nil
}
