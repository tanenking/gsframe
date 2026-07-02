package gsframe

import (
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/net/ws"
)

func StartWSServer(opt *gsinf.WebSocketServerConfig) gsinf.IWebSocketServer {
	return ws.CreateServer(opt)
}
