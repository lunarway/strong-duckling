package main

import (
	"testing"

	"github.com/lunarway/strong-duckling/internal/vici"
	"github.com/stretchr/testify/assert"
)

func TestCollectConnectionStats(t *testing.T) {
	tt := []struct {
		name  string
		conns []map[string]vici.IKEConf
	}{
		{
			name: "open connection",
			// config from temenos vpn in dev
			conns: []map[string]vici.IKEConf{
				{
					"gw-gw": vici.IKEConf{
						LocalAddrs:  []string{"10.0.224.131", "3.248.95.39"},
						RemoteAddrs: []string{"13.74.42.140"},
						LocalPort:   "",
						RemotePort:  "",
						Proposals:   []string(nil),
						Vips:        []string(nil),
						Version:     "IKEv2",
						Encap:       "",
						KeyingTries: "",
						RekeyTime:   "0",
						ReauthTime:  "10800",
						DPDDelay:    "",
						LocalAuth: vici.AuthConf{ID: "",
							Round:      "",
							AuthMethod: "",
							EAP_ID:     "",
							PubKeys:    []string(nil),
						},
						RemoteAuth: vici.AuthConf{
							ID:         "",
							Round:      "",
							AuthMethod: "",
							EAP_ID:     "",
							PubKeys:    []string(nil),
						},
						Pools: []string(nil),
						Children: map[string]vici.ChildSAConf{
							"net-net-0": vici.ChildSAConf{
								Local_ts:      []string(nil),
								Remote_ts:     []string(nil),
								Local_tso:     []string{"63.33.127.149/32"},
								Remote_tso:    []string{"10.110.47.128/27"},
								ESPProposals:  []string(nil),
								StartAction:   "",
								CloseAction:   "restart",
								ReqID:         "",
								RekeyTime:     "5400",
								ReplayWindow:  "",
								Mode:          "TUNNEL",
								InstallPolicy: "",
								UpDown:        "",
								Priority:      "",
								MarkIn:        "",
								MarkOut:       "",
								DpdAction:     "restart",
								LifeTime:      "",
								RekeyBytes:    "500000000",
								RekeyPackets:  "1000000"},
						},
						Mobike:      "",
						SendCertreq: "",
					},
				},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := collectConnectionStats(tc.conns, nil)
			assert.NoError(t, err, "unexpected error")
		})
	}
}

func TestCollectSasStats(t *testing.T) {
	tt := []struct {
		name string
		sas  []map[string]vici.IkeSa
	}{
		{
			name: "open connection",
			// config from temenos vpn in dev
			sas: []map[string]vici.IkeSa{
				{
					"gw-gw": vici.IkeSa{
						Uniqueid:        "65",
						Version:         "2",
						State:           "ESTABLISHED",
						Local_host:      "10.0.224.131",
						Local_port:      "4500",
						Local_id:        "3.248.95.39",
						Remote_host:     "13.74.42.140",
						Remote_port:     "4500",
						Remote_id:       "13.74.42.140",
						Remote_xauth_id: "",
						Initiator:       "yes",
						Initiator_spi:   "958a84787c5b2ca0",
						Responder_spi:   "999b0734c52826dd",
						Encr_alg:        "AES_CBC",
						Encr_keysize:    "256",
						Integ_alg:       "HMAC_SHA2_256_128",
						Integ_keysize:   "",
						Prf_alg:         "PRF_HMAC_SHA2_256",
						Dh_group:        "MODP_2048_256",
						Established:     "1768",
						Rekey_time:      "",
						Reauth_time:     "8479",
						Remote_vips:     []string(nil),
						Child_sas: map[string]vici.Child_sas{
							"net-net-0-176": vici.Child_sas{
								Reqid:         "65",
								State:         "INSTALLED",
								Mode:          "TUNNEL",
								Protocol:      "ESP",
								Encap:         "yes",
								Spi_in:        "cbe37e1d",
								Spi_out:       "e33400a8",
								Cpi_in:        "",
								Cpi_out:       "",
								Encr_alg:      "AES_CBC",
								Encr_keysize:  "256",
								Integ_alg:     "HMAC_SHA2_256_128",
								Integ_keysize: "",
								Prf_alg:       "",
								Dh_group:      "",
								Esn:           "",
								Bytes_in:      "17916",
								Packets_in:    "60",
								Use_in:        "407",
								Bytes_out:     "5146",
								Packets_out:   "52",
								Use_out:       "407",
								Rekey_time:    "3558",
								Life_time:     "4172",
								Install_time:  "1768",
								Local_ts:      []string{"63.33.127.149/32"},
								Remote_ts:     []string{"10.110.47.128/27"},
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := collectSasStats(tc.sas, nil)
			assert.NoError(t, err, "unexpected error")
		})
	}
}
