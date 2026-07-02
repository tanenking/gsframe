package kcpx

import (
	"time"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
)

type client struct {
	opt   *gsinf.KcpClientConfig
	conns chan *clientImpl
}

func CreateClient(opt *gsinf.KcpClientConfig) gsinf.IKcpClient {
	validateClientConfig(opt)
	var _client = &client{
		conns: make(chan *clientImpl, opt.PoolSize),
	}
	for range opt.PoolSize {
		_clientImpl := createClientImpl(_client)
		if _clientImpl == nil {
			continue
		}
		_client.conns <- _clientImpl
	}

	_client.start()
	return _client
}

func (r *client) start() {
	constants.Go(func() {
		defer constants.AutoRecover()()
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				r.healthCheck()
			default:
			}
		}
	})
}
func (r *client) healthCheck() {
}
