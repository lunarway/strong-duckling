package http

import (
	"net/http"

	"github.com/prometheus/common/log"
)

var (
	ready = true
)

func Define() *http.ServeMux {
	serveMux := http.NewServeMux()
	return serveMux
}

func Start(serveMux *http.ServeMux, listenAddress string) error {
	log.Infof("Listening on %s", listenAddress)
	return http.ListenAndServe(listenAddress, serveMux)
}

func Stop() {

}
