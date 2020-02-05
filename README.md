# strong-duckling

[![Build Status](https://travis-ci.com/lunarway/release-manager.svg?branch=master)](https://travis-ci.com/lunarway/strong-duckling)
[![Go Report Card](https://goreportcard.com/badge/github.com/lunarway/release-manager)](https://goreportcard.com/report/github.com/lunarway/strong-duckling)
[![GolangCI](https://raw.githubusercontent.com/golangci/golangci-web/master/src/assets/images/badge_a_plus_flat.svg)](https://golangci.com/r/github.com/lunarway/strong-duckling)

Strongswan sidecar and VPN tooling

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
