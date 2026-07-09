package kcpx

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
)

type client struct {
	opt        *gsinf.KcpClientConfig
	idgenerate int32
	ccount     int32
	maxcount   int32
	idx        int64
	connectors []*clientImpl
}

func CreateClient(opt *gsinf.KcpClientConfig) gsinf.IKcpClient {
	validateClientConfig(opt)

	var count = runtime.NumCPU() * 2

	var _client = &client{
		opt:        opt,
		maxcount:   int32(count),
		connectors: make([]*clientImpl, count),
	}
	for i := range count {
		cli := _client.createNewClientImpl()
		cli.reconnect()
		_client.connectors[i] = cli
	}

	_client.start()
	return _client
}

func (r *client) Send(header int64, msgID string, data []byte) error {
	defer constants.AutoRecover()()
	connector := r.GetConnector()
	if connector == nil {
		return fmt.Errorf("kcp client not found")
	}
	// defer r.putConnector(connector)
	return connector.Send(header, msgID, data)
}
func (r *client) SendControlMsg(cmd int32, msg string) error {
	var header = (int64(gsinf.KcpControlCMD) << int64(32)) | int64(cmd)
	return r.Send(header, msg, nil)
}

func (r *client) createNewClientImpl() *clientImpl {
	id := atomic.AddInt32(&r.idgenerate, 1)
	connector := createClientImpl(r, id)
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
func (r *client) GetConnector() *clientImpl {
	idx := atomic.AddInt64(&r.idx, 1)
	idx = idx % int64(r.maxcount)
	cli := r.connectors[idx]
	return cli
}
func (r *client) healthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		for idx, c := range r.connectors {
			if c == nil || !c.isValid() {
				r.connectors[idx] = r.createNewClientImpl()
			}
		}
	}
}

// func (r *client) getConnector() *clientImpl {
// 	fnget := func() *clientImpl {
// 		select {
// 		case connector := <-r.connectors:
// 			if connector.isValid() {
// 				return connector
// 			}
// 			connector.stop()
// 			atomic.AddInt32(&r.ccount, -1)
// 		default:
// 		}
// 		return nil
// 	}
// 	for {
// 		if connector := fnget(); connector != nil {
// 			return connector
// 		}
// 		select {
// 		case r.semaphore <- struct{}{}:
// 			//获取到创建许可
// 			return func() *clientImpl {
// 				defer func() { <-r.semaphore }()
// 				if atomic.LoadInt32(&r.ccount) < int32(r.opt.PoolSize) {
// 					return r.createNewClientImpl()
// 				}
// 				return nil
// 			}()
// 		case <-time.After(10 * time.Millisecond):
// 		}
// 	}
// }
// func (r *client) putConnector(connector *clientImpl) {
// 	if connector == nil {
// 		return
// 	}
// 	if !connector.isValid() {
// 		connector.stop()
// 		atomic.AddInt32(&r.ccount, -1)
// 		return
// 	}
// 	select {
// 	case r.connectors <- connector:
// 		//归还成功
// 	default:
// 		//池已满,关闭连接
// 		connector.stop()
// 		atomic.AddInt32(&r.ccount, -1)
// 	}
// }
