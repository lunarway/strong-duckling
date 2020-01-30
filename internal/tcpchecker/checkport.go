package tcpchecker

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
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

	output := strings.Builder{}
	for scanner.Scan() {
		fmt.Fprintf(&output, "%s\n", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		var netError net.Error
		if errors.As(err, &netError) && netError.Timeout() {
			reporter.ReportPortCheck(Report{
				Address: address,
				Port:    port,
				Open:    true,
				Status:  "Open",
				Error:   nil,
				Content: output.String(),
			})
			return output.String(), nil
		}
		reporter.ReportPortCheck(Report{
			Address: address,
			Port:    port,
			Open:    false,
			Status:  "Scanner error",
			Error:   err,
			Content: output.String(),
		})
		return output.String(), err
	}
	reporter.ReportPortCheck(Report{
		Address: address,
		Port:    port,
		Open:    true,
		Status:  "Open",
		Error:   nil,
		Content: output.String(),
	})
	return output.String(), nil
}
