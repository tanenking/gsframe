package ws

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader websocket.Upgrader
)

func checkOrigin(r *http.Request) bool {
	return true
}

func init() {
	upgrader = websocket.Upgrader{
		HandshakeTimeout: config.WriteTimeout,
		ReadBufferSize:   int(config.GoReadWriteBufferSize),
		WriteBufferSize:  int(config.GoReadWriteBufferSize),
		CheckOrigin:      checkOrigin,
		WriteBufferPool: &sync.Pool{
			New: func() any { return make([]byte, 0, config.MaxPacketSize) },
		},
	}
}
