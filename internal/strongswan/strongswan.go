package strongswan

import (
	"fmt"

	"github.com/lunarway/strong-duckling/internal/vici"
	"github.com/prometheus/common/log"
)

type Reporter interface {
	IKESAStatus(ikeName string, conn vici.IKEConf, sa *vici.IkeSa)
}

func Collect(client *vici.ClientConn, reporter Reporter) {
	conns, err := connections(client)
	if err != nil {
		log.Errorf("Failed to get strongswan connections: %v", err)
		return
	}
	sas, err := ikeSas(client)
	if err != nil {
		log.Errorf("Failed to get strongswan sas: %v", err)
		return
	}
	collectSasStats(conns, sas, reporter)
}

func connections(client *vici.ClientConn) ([]map[string]vici.IKEConf, error) {
	connList, err := client.ListConns("")
	if err != nil {
		return nil, fmt.Errorf("list vici conns: %w", err)
	}
	return connList, nil
}

func ikeSas(client *vici.ClientConn) ([]map[string]vici.IkeSa, error) {
	sasList, err := client.ListSas("", "")
	if err != nil {
		return nil, fmt.Errorf("list vici sas: %w", err)
	}
	return sasList, nil
}

func collectSasStats(configs []map[string]vici.IKEConf, sas []map[string]vici.IkeSa, reporter Reporter) {
	/*
		if connection is configured and ikesa is missing somethings wrong, so we
		track the expected connections and the actual ones and can the report if
		something is missing after looping through the child SAs.
	*/
	expectedConnections := make(map[string]vici.IKEConf)
	for _, conf := range configs {
		for name := range conf {
			expectedConnections[name] = conf[name]
		}
	}

	for _, sa := range sas {
		for ikeName, ikeSa := range sa {
			conf, ok := expectedConnections[ikeName]
			if !ok {
				log.Errorf("Unexpected SA: %s: %#v", ikeName, ikeSa)
				continue
			}
			log.With("conf", conf).
				With("sa", ikeSa).
				With("ikeName", ikeName).
				Infof("Reporting on ike name '%s'", ikeName)
			reporter.IKESAStatus(ikeName, conf, &ikeSa)
			delete(expectedConnections, ikeName)
		}
	}
	for ikeName, conf := range expectedConnections {
		log.With("conf", conf).
			With("ikeName", ikeName).
			Infof("Reporting config without active SAs on '%s'", ikeName)
		reporter.IKESAStatus(ikeName, conf, nil)
	}
}
