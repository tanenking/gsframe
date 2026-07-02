package gsframe

import (
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/net/ws"
)

func StartWSServer(opt *gsinf.WebSocketServerConfig) gsinf.IWebSocketServer {
	return ws.CreateServer(opt)
}

// type TcpInitOption struct {
// 	TCPPort        int
// 	MaxConn        int
// 	MaxPacketSize  int
// 	IsWebSocket    bool
// 	DirectProtoMsg bool

// 	OnServerPreStart func(server gsinf.IServer)

// 	OnConnStart func(conn gsinf.IConnection)
// 	OnConnStop  func(conn gsinf.IConnection)
// }

// func GetTcpGlobalConfig() *gsinf.TcpGlobalConfig_t {
// 	return zcommon.GlobalObject
// }

// func StartTcpServer(opt TcpInitOption) gsinf.IServer {
// 	if opt.TCPPort <= 0 || opt.TCPPort >= 65535 {
// 		return nil
// 	}
// 	zcommon.GlobalObject.TCPPort = opt.TCPPort
// 	zcommon.GlobalObject.MaxConn = opt.MaxConn
// 	zcommon.GlobalObject.MaxPacketSize = uint32(opt.MaxPacketSize)
// 	zcommon.GlobalObject.DirectProtoMsg = opt.DirectProtoMsg

// 	var server gsinf.IServer
// 	if opt.IsWebSocket {
// 		if opt.DirectProtoMsg {
// 			server = zwspb.NewServer()
// 		} else {
// 			server = zws.NewServer()
// 		}
// 	} else {
// 		server = znet.NewServer()
// 	}
// 	if opt.OnServerPreStart != nil {
// 		opt.OnServerPreStart(server)
// 	}
// 	server.SetOnConnStart(opt.OnConnStart)
// 	server.SetOnConnStop(opt.OnConnStop)

// 	constants.Go(func() { server.Serve() })

// 	return server
// }

// func SendGroupMsg(groupName string, msgID string, data []byte) {
// 	zcommon.SendGroupMsg(groupName, msgID, data)
// }
