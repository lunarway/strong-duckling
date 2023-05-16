package http

import (
	"net/http"

	"go.uber.org/zap"
)

func Define() *http.ServeMux {
	serveMux := http.NewServeMux()
	return serveMux
}

func Start(serveMux *http.ServeMux, listenAddress string) error {
	zap.L().Sugar().Infof("Listening on %s", listenAddress)
	return http.ListenAndServe(listenAddress, serveMux)
}
