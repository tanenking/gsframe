package zws

import (
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/tanenking/gsframe/internal/tcp/zcommon"
)

var (
	upgrader websocket.Upgrader
	cID      atomic.Uint32
)

func checkOrigin(r *http.Request) bool {
	return true
}

func init() {
	upgrader = websocket.Upgrader{
		HandshakeTimeout: zcommon.GlobalObject.WriteTimeout,
		ReadBufferSize:   int(zcommon.GlobalObject.MaxPacketSize), //4096,
		WriteBufferSize:  int(zcommon.GlobalObject.MaxPacketSize), //4096,
		CheckOrigin:      checkOrigin,
		WriteBufferPool: &sync.Pool{
			New: func() any { return make([]byte, 0, zcommon.GlobalObject.MaxPacketSize) },
		},
	}
	cID.Store(0)
}
