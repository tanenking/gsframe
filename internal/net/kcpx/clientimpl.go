package kcpx

import (
	"context"
	"time"

	"github.com/tanenking/gsframe/internal/logger"
	"github.com/xtaci/kcp-go/v5"
)

type clientImpl struct {
	_client *client
	ctx     context.Context
	cancel  context.CancelFunc
	conn    *kcp.UDPSession
}

func createClientImpl(_client *client) *clientImpl {
	impl := &clientImpl{
		_client: _client,
	}
	impl.ctx, impl.cancel = context.WithCancel(context.Background())

	conn, err := kcp.Dial(_client.opt.Address)
	if err != nil {
		logger.Log().Error(`Dial err %+v`, err)
		return nil
	}
	impl.conn = conn.(*kcp.UDPSession)

	impl.conn.SetReadDeadline(time.Now().Add(impl._client.opt.ReadTimeout))
	impl.conn.SetWriteDeadline(time.Now().Add(impl._client.opt.WriteTimeout))

	impl.conn.SetRateLimit(0)
	impl.conn.SetStreamMode(impl._client.opt.StreamMode)
	impl.conn.SetNoDelay(1, 10, 2, 1)
	impl.conn.SetACKNoDelay(impl._client.opt.NoDelay)
	impl.conn.SetWindowSize(1024, 1024)
	impl.conn.SetMtu(1400)

	impl.conn.SetReadBuffer(int(impl._client.opt.TcpReadWriteBufferSize))
	impl.conn.SetWriteBuffer(int(impl._client.opt.TcpReadWriteBufferSize))

	return impl
}
