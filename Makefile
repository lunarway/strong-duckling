
STRONGSWAN_DOCKER_IMAGE="philplckthun/strongswan"
VICI_SOCKET="/tmp/strongswan"

.PHONY: run/vpn

run/vpn:
	docker run -it --rm -e VPN_PSK="psk" -e VPN_PASSWORD="pass" -v $(VICI_SOCKET):/run -p 500:500/udp -p 4500:4500/udp -p 1701:1701/udp -p 4502:4502 --privileged ${STRONGSWAN_DOCKER_IMAGE}

swanctl/list-conns:
	docker run -it --rm -v $(VICI_SOCKET):/run $(STRONGSWAN_DOCKER_IMAGE) swanctl --list-conns

swanctl/list-sas:
	docker run -it --rm -v $(VICI_SOCKET):/run $(STRONGSWAN_DOCKER_IMAGE) swanctl --list-sas

DATE:=`date +%Y-%m-%d\_%H:%M`
GIT_SHA:=`git rev-parse HEAD`

build:
	GOOS=linux go build -o strong-duckling -ldflags="-X main.version=$(GIT_SHA)_$(DATE)" main.go

MOCKERY_ARGS=-case=underscore -inpkg -testonly
generate/mock:
	go get github.com/vektra/mockery/.../
	mockery $(MOCKERY_ARGS) -dir internal/stats -name .*Reporter
