package gsframe

import (
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/net/kcpx"
)

func StartKcpServer(opt *gsinf.KcpServerConfig) gsinf.IKcpServer {
	return kcpx.CreateServer(opt)
}

func CreateKcpClient(opt *gsinf.KcpClientConfig) gsinf.IKcpClient {
	return kcpx.CreateClient(opt)
}
