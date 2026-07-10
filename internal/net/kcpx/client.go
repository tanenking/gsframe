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
	invalids   chan *clientImpl
}

func CreateClient(opt *gsinf.KcpClientConfig) gsinf.IKcpClient {
	validateClientConfig(opt)

	var count = runtime.NumCPU() * 2

	var _client = &client{
		opt:        opt,
		maxcount:   int32(count),
		connectors: make([]*clientImpl, count),
		invalids:   make(chan *clientImpl, count),
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
	connector := r.getConnector()
	if connector == nil || !connector.isValid() {
		return fmt.Errorf("kcp client not found")
	}
	// defer r.putConnector(connector)
	return connector.Send(header, msgID, data)
}
func (r *client) SendControlMsg(cmd int32, msg string) error {
	var header = (int64(gsinf.KcpControlCMD) << int64(32)) | int64(cmd)
	return r.Send(header, msg, nil)
}
func (r *client) GetConnector(flag int32) gsinf.IKcpClientImpl {
	idx := flag % r.maxcount
	connector := r.connectors[idx]
	if connector == nil || !connector.isValid() {
		return nil
	}
	return connector
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

func (r *client) getConnector() *clientImpl {
	idx := atomic.AddInt64(&r.idx, 1)
	idx = idx % int64(r.maxcount)
	connector := r.connectors[idx]
	return connector
}

func (r *client) onImplError(impl *clientImpl) {
	defer constants.AutoRecover()()
	r.invalids <- impl
}

func (r *client) healthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	defer func() {
		ticker.Stop()
		close(r.invalids)
		constants.AppExitDone()
	}()
	constants.AppExitWait()
	for {
		select {
		case <-constants.ExitChannel:
			return
		case c, ok := <-r.invalids:
			if !ok {
				return
			}
			if !c.isValid() {
				c.reconnect()
			}
		case <-ticker.C:
			for idx, c := range r.connectors {
				if c == nil {
					r.connectors[idx] = r.createNewClientImpl()
				}
				if !c.isValid() {
					c.reconnect()
				}
			}
		}
	}
}
