package util_http

import (
	"net/http"
)

var (
	httpServer *http.Server
)

func init() {
	httpServer = nil
}
