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
	semaphore  chan struct{} //最大并发创建数
}

func CreateClient(opt *gsinf.KcpClientConfig) gsinf.IKcpClient {
	validateClientConfig(opt)

	if opt.PoolSize > 200 {
		opt.PoolSize = 200
	}
	var chancount = opt.PoolSize

	var _client = &client{
		opt:        opt,
		connectors: make(chan *clientImpl, chancount),
		semaphore:  make(chan struct{}, 1),
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
	fnget := func() *clientImpl {
		select {
		case connector := <-r.connectors:
			if connector.isValid() {
				return connector
			}
			connector.stop()
			atomic.AddInt32(&r.ccount, -1)
		default:
		}
		return nil
	}
	for {
		if connector := fnget(); connector != nil {
			return connector
		}
		select {
		case r.semaphore <- struct{}{}:
			//获取到创建许可
			return func() *clientImpl {
				defer func() { <-r.semaphore }()
				if atomic.LoadInt32(&r.ccount) < int32(r.opt.PoolSize) {
					return r.createNewClientImpl()
				}
				return nil
			}()
		case <-time.After(1 * time.Millisecond):
		}
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
		//归还成功
	default:
		//池已满,关闭连接
		connector.stop()
		atomic.AddInt32(&r.ccount, -1)
	}
}
