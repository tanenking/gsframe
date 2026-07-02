package kcpx

import (
	"github.com/tanenking/gsframe/gsinf"
)

type client struct {
	address string
	pool    chan *clientImpl
}

func CreateClient(opt *gsinf.KcpClientConfig) gsinf.IKcpClient {
	validateClientConfig(opt)
	var _client = &client{
		address: opt.Address,
		pool:    make(chan *clientImpl, opt.PoolSize),
	}

	return _client
}
