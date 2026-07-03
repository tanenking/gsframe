package kcpx

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
)

type client struct {
	opt        *gsinf.KcpClientConfig
	cindex     int32
	ccount     int32
	connectors chan *clientImpl
}

func CreateClient(opt *gsinf.KcpClientConfig) gsinf.IKcpClient {
	validateClientConfig(opt)

	var chancount = opt.PoolSize
	if chancount < 1024 {
		chancount = 1024
	}

	var _client = &client{
		connectors: make(chan *clientImpl, chancount),
	}
	for range opt.PoolSize {
		_clientImpl := _client.createNewClientImpl()
		if _clientImpl == nil {
			continue
		}
		_client.connectors <- _clientImpl
	}

	_client.start()
	return _client
}

func (r *client) Send(header int64, msgID string, data []byte) error {
	defer constants.AutoRecover()()
	connector := r.getConnector()
	if connector == nil {
		return fmt.Errorf("kcp client not found")
	}
	defer r.putConnector(connector)
	return connector.send(header, msgID, data)
}

func (r *client) createNewClientImpl() *clientImpl {
	index := atomic.AddInt32(&r.cindex, 1)
	connector := createClientImpl(r, index)
	if connector != nil {
		atomic.AddInt32(&r.ccount, 1)
	}

	return connector
}

func (r *client) start() {
	constants.Go(func() {
		defer constants.AutoRecover()()
		r.healthCheck()
	})
}
func (r *client) healthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		// 检查连接
		var _break = false
		for !_break {
			select {
			case connector, ok := <-r.connectors:
				if !ok {
					return
				}
				if !connector.isValid() {
					connector.stop()
					atomic.AddInt32(&r.ccount, -1)
				} else {
					r.putConnector(connector)
				}
			default:
				_break = true
			}
		}
	}
}
func (r *client) getConnector() *clientImpl {
	select {
	case connector := <-r.connectors:
		// 检查连接是否有效
		if connector.isValid() {
			return connector
		}
		return r.createNewClientImpl()
	case <-time.After(100 * time.Millisecond):
		// 池空且未超时，尝试新建
		return r.createNewClientImpl()
	}
}
func (r *client) putConnector(connector *clientImpl) {
	if connector == nil {
		return
	}
	if !connector.isValid() {
		connector.stop()
		atomic.AddInt32(&r.ccount, -1)
		return
	}
	select {
	case r.connectors <- connector:
		// 归还成功
	default:
		// 池已满，关闭连接
		connector.stop()
		atomic.AddInt32(&r.ccount, -1)
	}
}
