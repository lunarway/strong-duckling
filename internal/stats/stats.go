package stats

import (
	"fmt"

	"github.com/lunarway/strong-duckling/internal/vici"
	"github.com/prometheus/common/log"
)

type Reporter interface {
	IKEConnectionConfiguration(string, vici.IKEConf)
	IKESAStatus(conn vici.IKEConf, sa *vici.IkeSa)
}

func Collect(client *vici.ClientConn, reporter Reporter) {
	conns, err := connections(client)
	if err != nil {
		log.Errorf("Failed to get strongswan connections: %v", err)
		return
	}
	collectConnectionStats(conns, reporter)

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

func collectConnectionStats(conns []map[string]vici.IKEConf, reporter Reporter) {
	log.Infof("Connections: %d", len(conns))
	for _, connection := range conns {
		for ikeName, ike := range connection {
			log.Infof("  ikeName: %s: ike: %#v", ikeName, ike)
			reporter.IKEConnectionConfiguration(ikeName, ike)
		}
	}
}

func collectSasStats(configs []map[string]vici.IKEConf, sas []map[string]vici.IkeSa, reporter Reporter) {
	/*
		if connection is configured and ikesa is missing somethings wrong
	*/
	expectedConnections := make(map[string]vici.IKEConf)
	for _, conf := range configs {
		for name := range conf {
			expectedConnections[name] = conf[name]
		}
	}

	log.Infof("Sas: %d", len(sas))
	for _, sa := range sas {
		for ikeName, ikeSa := range sa {
			conf, ok := expectedConnections[ikeName]
			if !ok {
				log.Errorf("Unexpected SA: %s: %#v", ikeName, ikeSa)
				continue
			}
			log.Infof("  ikeName: %s: sa: %#v", ikeName, ikeSa)
			reporter.IKESAStatus(conf, &ikeSa)
			delete(expectedConnections, ikeName)
		}
	}
	for _, conf := range expectedConnections {
		reporter.IKESAStatus(conf, nil)
	}
}

type infoReporter interface {
	Info(buildVersion string)
}

func RunningVersion(version string, reporter infoReporter) error {
	log.Infof("Strong duckling version %s", version)
	reporter.Info(version)
	return nil
}
