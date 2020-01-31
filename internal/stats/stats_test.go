package stats

import (
	"testing"

	"github.com/lunarway/strong-duckling/internal/vici"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// func TestCollectConnectionStats(t *testing.T) {
// 	tt := []struct {
// 		name  string
// 		conns []map[string]vici.IKEConf
// 	}{
// 		{
// 			name: "open connection",
// 			// config from temenos vpn in dev
// 			conns: []map[string]vici.IKEConf{
// 				{
// 					"gw-gw": vici.IKEConf{
// 						LocalAddresses:    []string{"10.0.224.131", "3.248.95.39"},
// 						RemoteAddresses:   []string{"13.74.42.140"},
// 						LocalPort:         "",
// 						RemotePort:        "",
// 						Proposals:         []string(nil),
// 						VIPs:              []string(nil),
// 						IKEVersion:        "IKEv2",
// 						Encapsulation:     "",
// 						KeyingTries:       "",
// 						RekeyTimeSeconds:  "0",
// 						ReauthTimeSeconds: "10800",
// 						DPDDelay:          "",
// 						Pools:             []string(nil),
// 						Children: map[string]vici.ChildSAConf{
// 							"net-net-0": vici.ChildSAConf{
// 								LocalTrafficSelectors:  []string{"63.33.127.149/32"},
// 								RemoteTrafficSelectors: []string{"10.110.47.128/27"},
// 								ESPProposals:           []string(nil),
// 								StartAction:            "",
// 								CloseAction:            "restart",
// 								ReqID:                  "",
// 								RekeyTimeSeconds:       "5400",
// 								ReplayWindow:           "",
// 								IPsecMode:              "TUNNEL",
// 								InstallPolicy:          "",
// 								UpDown:                 "",
// 								Priority:               "",
// 								MarkIn:                 "",
// 								MarkOut:                "",
// 								DpdAction:              "restart",
// 								LifeTime:               "",
// 								RekeyBytes:             "500000000",
// 								RekeyPackets:           "1000000"},
// 						},
// 						MOBIKE:          "",
// 						SendCertRequest: "",
// 					},
// 				},
// 			},
// 		},
// 	}
// 	for _, tc := range tt {
// 		t.Run(tc.name, func(t *testing.T) {
// 			reporter := MockReporter{}
// 			reporter.Test(t)
// 			collectConnectionStats(tc.conns)
// 		})
// 	}
// }

// {
// 	name: "open connection",
// 	// config from temenos vpn in dev
// 	sas: []map[string]vici.IkeSa{
// 		{
// 			"gw-gw": vici.IkeSa{
// 				UniqueID:            "67",
// 				IKEVersion:          "2",
// 				State:               "ESTABLISHED",
// 				LocalHost:           "10.0.224.131",
// 				LocalPort:           "4500",
// 				LocalID:             "3.248.95.39",
// 				RemoteHost:          "13.74.42.140",
// 				RemotePort:          "4500",
// 				RemoteID:            "13.74.42.140",
// 				RemoteXAuthID:       "",
// 				RemoteEAPID:         "",
// 				Initiator:           "yes",
// 				InitiatorSPI:        "020d4f1dd06ec915",
// 				ResponderSPI:        "7de60890aeae0dc5",
// 				EncryptionAlgorithm: "AES_CBC",
// 				EncryptionKeySize:   "256",
// 				IntegrityAlgorithm:  "HMAC_SHA2_256_128",
// 				IntegrityKeySize:    "",
// 				PRFAlgorithm:        "PRF_HMAC_SHA2_256",
// 				DHGroup:             "MODP_2048_256",
// 				EstablishedSeconds:  "3797",
// 				RekeyTimeSeconds:    "",
// 				ReauthTimeSeconds:   "6545",
// 				LocalVIPs:           []string(nil),
// 				RemoteVIPs:          []string(nil),
// 				ChildSAs: map[string]vici.ChildSA{"net-net-0-181": vici.ChildSA{Name: "net-net-0",
// 					UniqueID:               "181",
// 					ReqID:                  "67",
// 					State:                  "INSTALLED",
// 					IPsecMode:              "TUNNEL",
// 					IPsecProtocol:          "ESP",
// 					UDPEncapsulation:       "yes",
// 					SPIIn:                  "c10ead54",
// 					SPIOut:                 "57908b22",
// 					CPIIn:                  "",
// 					CPIOut:                 "",
// 					EncryptionAlgorithm:    "AES_CBC",
// 					EncryptionKeySize:      "256",
// 					IntegrityAlgorithm:     "HMAC_SHA2_256_128",
// 					IntegrityKeySize:       "",
// 					PRFAlgorithm:           "",
// 					DHGroup:                "",
// 					ExtendedSequenceNumber: "",
// 					BytesIn:                "35832",
// 					BytesOut:               "10240",
// 					PacketsIn:              "120",
// 					PacketsOut:             "103",
// 					LastPacketInSeconds:    "378",
// 					LastPacketOutSeconds:   "378",
// 					RekeyTimeSeconds:       "1299",
// 					LifeTimeSeconds:        "2143",
// 					InstallTimeSeconds:     "3797",
// 					LocalTrafficSelectors:  []string{"63.33.127.149/32"},
// 					RemoteTrafficSelectors: []string{"10.110.47.128/27"},
// 				},
// 				},
// 			},
// 		},
// 	},
// },
func TestCollectSasStats(t *testing.T) {
	tt := []struct {
		name              string
		connectionConfigs []map[string]vici.IKEConf
		sas               []map[string]vici.IkeSa
		conf              *vici.IKEConf
		sa                *vici.IkeSa
	}{
		{
			name: "connection missing from config",
			connectionConfigs: []map[string]vici.IKEConf{
				{
					"gw-gw": vici.IKEConf{
						Unique: "id-1",
					},
				},
			},
			conf: &vici.IKEConf{
				Unique: "id-1",
			},
			sa: nil,
		},
		{
			name: "connection missing from config",
			connectionConfigs: []map[string]vici.IKEConf{
				{
					"gw-gw": vici.IKEConf{
						Unique: "id-1",
					},
				},
			},
			conf: &vici.IKEConf{
				Unique: "id-1",
			},
			sa: nil,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			reporter := MockReporter{}
			reporter.Test(t)
			var reportedConf *vici.IKEConf
			var reportedSA *vici.IkeSa
			reporter.On("IKESAStatus", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				conf, sa := args[0].(vici.IKEConf), args[1].(*vici.IkeSa)
				reportedConf = &conf
				reportedSA = sa
			})

			collectSasStats(tc.connectionConfigs, tc.sas, &reporter)

			assert.Equal(t, tc.conf, reportedConf, "IKE Conf not as expected")
			assert.Equal(t, tc.sa, reportedSA, "IKE SA not as expected")
		})
	}
}
