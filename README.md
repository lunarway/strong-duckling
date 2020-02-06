# strong-duckling

[![Build Status](https://travis-ci.com/lunarway/release-manager.svg?branch=master)](https://travis-ci.com/lunarway/strong-duckling)
[![Go Report Card](https://goreportcard.com/badge/github.com/lunarway/release-manager)](https://goreportcard.com/report/github.com/lunarway/strong-duckling)
[![GolangCI](https://golangci.com/badges/github.com/lunarway/strong-duckling.svg)](https://golangci.com/r/github.com/lunarway/strong-duckling)

Strongswan sidecar and VPN tooling

# Metrics

Enable HTTP server exposing prometheus metrics by setting `--listen` to a port, e.g. `--listen=:9100`.
The application exposes Prometheus metrics on `/metrics` for general insight into the application along with other features if enabled.

| Name                   | Labels                         | Description                                               |
| ---------------------- | ------------------------------ | --------------------------------------------------------- |
| `strong_duckling_info` | `version`,`strongswan_version` | Metadata such as version info of the application it self. |

## IKE SA metrics

Enable Strongswan metrics by setting `--vici-socket` to a charon socket of a running strongswan process.
Usually this is `/var/run/charon.vici`.

| Name                                                          | Type      | Labels | Description                              |
| ------------------------------------------------------------- | --------- | ------ | ---------------------------------------- |
| `strong_duckling_ike_sa_established_seconds`                  | Gauge     |        | Time the SA have been established        |
| `strong_duckling_ike_sa_packets_in_total`                     | Counter   |        | Total number of received packets         |
| `strong_duckling_ike_sa_packets_out_total`                    | Counter   |        | Total number of transmitted packets      |
| `strong_duckling_ike_sa_packets_in_silence_duration_seconds`  | Histogram |        | Duration of silences between packets in  |
| `strong_duckling_ike_sa_packets_out_silence_duration_seconds` | Histogram |        | Duration of silences between packets out |
| `strong_duckling_ike_sa_bytes_in_total`                       | Counter   |        | Total number of received bytes           |
| `strong_duckling_ike_sa_bytes_out_total`                      | Counter   |        | Total number of transmitted bytes        |
| `strong_duckling_ike_sa_installs_total`                       | Counter   |        | Total number of SA installs              |
| `strong_duckling_ike_sa_rekey_seconds`                        | Histogram |        | Duration between re-keying               |
| `strong_duckling_ike_sa_lifetime_seconds`                     | Histogram |        | Duration of child SA connections         |
| `strong_duckling_ike_sa_state_info`                           | Gauge     |        | Metadata on the state of the SA          |
| `strong_duckling_ike_sa_child_state_info`                     | Gauge     |        | Metadata on the state of the child SA    |

## Local development setup
To use the test setup start a linux build watcher (requires nodemon) like this:

```bash
./build-linux.sh
```

In a separate terminal start the docker-compose configuration:

```bash
docker-compose up -d
```

This will start 2 linked docker containers each running:

* StrongSwan VPN
* A small nodejs HTTP server on :8080
* strong-duckling

The setup is configured to automatically connect the 2 containers using StrongSwan through an IKE v2 tunnel. The machines have added internal IPs `10.101.0.1` and `10.102.0.1`.
