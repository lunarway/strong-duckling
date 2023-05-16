package strongswan

import (
	"fmt"

	"github.com/lunarway/strong-duckling/internal/vici"
	"go.uber.org/zap"
)

func Collect(client *vici.ClientConn, ikeSAStatusReceivers []IKESAStatusReceiver) {
	conns, err := connections(client)
	if err != nil {
		zap.L().Sugar().Errorf("Failed to get strongswan connections: %v", err)
		return
	}
	sas, err := ikeSas(client)
	if err != nil {
		zap.L().Sugar().Errorf("Failed to get strongswan sas: %v", err)
		return
	}
	collectSasStats(conns, sas, ikeSAStatusReceivers)
}

func connections(client *vici.ClientConn) (map[string]vici.IKEConf, error) {
	connList, err := client.ListConns("")
	if err != nil {
		return nil, fmt.Errorf("list vici conns: %w", err)
	}
	return connList, nil
}

func ikeSas(client *vici.ClientConn) (map[string]vici.IkeSa, error) {
	sasList, err := client.ListSas("", "")
	if err != nil {
		return nil, fmt.Errorf("list vici sas: %w", err)
	}
	return sasList, nil
}

func collectSasStats(configs map[string]vici.IKEConf, sas map[string]vici.IkeSa, ikeSAStatusReceivers []IKESAStatusReceiver) {
	ikeNames := make(map[string]struct{})
	for ikeName := range configs {
		ikeNames[ikeName] = struct{}{}
	}
	for ikeName := range sas {
		ikeNames[ikeName] = struct{}{}
	}

	var ikeSAStatuses []IKESAStatus
	for ikeName := range ikeNames {
		config, configFound := configs[ikeName]
		ikeSA, ikeSAFound := sas[ikeName]
		switch {
		case configFound && ikeSAFound:
			ikeSAStatuses = append(ikeSAStatuses, mapToIKESAStatus(ikeName, config, &ikeSA))
		case configFound && !ikeSAFound:
			ikeSAStatuses = append(ikeSAStatuses, mapToIKESAStatus(ikeName, config, nil))
		case !configFound && ikeSAFound:
			zap.L().Sugar().Errorf("Unexpected IKE_SA Status for IKE Name %s: %#v", ikeName, ikeSA)
		}
	}

	for _, ikeSAStatus := range ikeSAStatuses {
		for _, reporter := range ikeSAStatusReceivers {
			reporter.IKESAStatus(ikeSAStatus)
		}
	}
}

func mapToIKESAStatus(ikeName string, config vici.IKEConf, ikeSA *vici.IkeSa) IKESAStatus {
	status := IKESAStatus{
		Name:          ikeName,
		Configuration: config,
		State:         ikeSA,
	}

	childNames := make(map[string]struct{})
	for childName := range config.Children {
		childNames[childName] = struct{}{}
	}
	if ikeSA != nil {
		for _, childSA := range ikeSA.ChildSAs {
			childNames[childSA.Name] = struct{}{}
		}
	}

	for childName := range childNames {
		childConfig, childConfigFound := config.Children[childName]
		childSA, childSAFound := findMatchingChildSA(ikeSA, childName)

		switch {
		case childConfigFound && childSAFound:
			status.ChildSA = append(status.ChildSA, ChildSAStatus{
				Name:          childName,
				Configuration: childConfig,
				State:         &childSA,
			})
		case childConfigFound && !childSAFound:
			status.ChildSA = append(status.ChildSA, ChildSAStatus{
				Name:          childName,
				Configuration: childConfig,
			})
		case !childConfigFound && childSAFound:
			zap.L().Sugar().Errorf("Unexpected CHILD_SA Status for IKE Name %s and Child SA Name %s: %#v", ikeName, childName, ikeSA)
		}
	}
	return status
}

func findMatchingChildSA(ikeSA *vici.IkeSa, childName string) (vici.ChildSA, bool) {
	if ikeSA == nil {
		return vici.ChildSA{}, false
	}

	for _, c := range ikeSA.ChildSAs {
		if c.Name == childName {
			return c, true
		}
	}
	return vici.ChildSA{}, false
}
