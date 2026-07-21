package kcpx

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/logger"
	"github.com/tanenking/gsframe/internal/net/common"
	"github.com/xtaci/kcp-go/v5"
)

type clientImpl struct {
	_client         *client
	ctx             context.Context
	cancel          context.CancelFunc
	conn            *kcp.UDPSession
	connId          int32
	closed          int32
	running         int32
	writeBufferList chan *common.Message
}

func createClientImpl(_client *client, connId int32) *clientImpl {
	impl := &clientImpl{
		_client: _client,
		connId:  connId,
		running: 0,
	}

	return impl
}
func (c *clientImpl) reconnect() {
	defer constants.AutoRecover()()

	if !atomic.CompareAndSwapInt32(&c.running, 0, 1) {
		return
	}
	conn, err := kcp.Dial(c._client.opt.Address)
	if err != nil {
		logger.Log().Error(`clientImpl %d Dial err %+v`, c.connId, err)
		atomic.StoreInt32(&c.running, 0)
		return
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.closed = 0

	c.writeBufferList = make(chan *common.Message, c._client.opt.WriteMessageBufferLen)

	c.conn = conn.(*kcp.UDPSession)

	c.conn.SetReadDeadline(time.Time{})
	c.conn.SetWriteDeadline(time.Now().Add(c._client.opt.WriteTimeout))

	c.conn.SetRateLimit(0)
	c.conn.SetStreamMode(c._client.opt.StreamMode)
	c.conn.SetNoDelay(1, 10, 2, 1)
	c.conn.SetACKNoDelay(true)
	c.conn.SetWindowSize(1024, 1024)
	c.conn.SetMtu(1472)
	c.conn.SetWriteDelay(true)

	c.conn.SetReadBuffer(int(c._client.opt.TcpReadWriteBufferSize))
	c.conn.SetWriteBuffer(int(c._client.opt.TcpReadWriteBufferSize))

	constants.Go(func() { c.start() })
}

// Stop 停止连接，结束当前连接状态M
func (c *clientImpl) stop() {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return
	}
	constants.Go(func() {
		c.cancel()
	})
}
func (c *clientImpl) isValid() bool {
	return atomic.LoadInt32(&c.running) > 0 && atomic.LoadInt32(&c.closed) == 0 && c.conn != nil
}

func (c *clientImpl) Send(header int64, msgID string, data []byte) error {
	defer constants.AutoRecover()()
	if atomic.LoadInt32(&c.closed) > 0 {
		return fmt.Errorf("kcp client %d closed when send buff msg", c.connId)
	}

	var msg = common.CreateMessage(c._client.opt.ByteOrder)
	msg.ID = msgID
	msg.Header = header
	if cap(msg.Data) < len(data) {
		msg.Data = make([]byte, 0, len(data))
	}
	msg.Data = msg.Data[:len(data)]
	copy(msg.Data, data)

	c.writeBufferList <- msg

	return nil
}
func (c *clientImpl) SendControlMsg(cmd int32, msg string) error {
	var header = (int64(gsinf.KcpControlCMD) << int64(32)) | int64(cmd)
	return c.Send(header, msg, nil)
}

func (c *clientImpl) start() {
	if c._client.opt.OnConnectorCreate != nil {
		c._client.opt.OnConnectorCreate(c)
	}
	constants.Go(func() {
		c.startReader()
	})
	constants.Go(func() {
		c.startWriter()
	})

	<-c.ctx.Done()

	//关闭该链接全部管道
	close(c.writeBufferList)

	//关闭连接,将还未发送完的数据发完
	c.sendRest()

	c.finalizer()
}

func (c *clientImpl) readMessage(bs *common.ByteBuffer) (err error) {
	if c.conn == nil {
		return errors.New(`kcp client readMessage conn is nil`)
	}
	bs.Data = bs.Data[:0]
	var recvtotal = 0
	var recv = make([]byte, cap(bs.Data))
	for {
		recv = recv[:0]
		var recvcount int
		recvcount, err = c.conn.Read(recv)
		if err != nil {
			logger.Log().Error("kcp client read error %v", err)
			return
		}
		if recvcount == 0 {
			return
		}
		if cap(bs.Data) < (recvtotal + recvcount) {
			//扩容
			newBuf := make([]byte, len(bs.Data), cap(bs.Data)*2)
			copy(newBuf, bs.Data)
			bs.Data = newBuf
		}
		copy(bs.Data[recvtotal:], recv)
		recvtotal += recvcount

		totallen := common.ReadMessageTotalLength(bs.Data, config.ByteOrder)
		if recvtotal >= int(totallen) {
			break
		}
	}
	return
}

// StartReader 读消息Goroutine，用于从客户端中读取数据
func (c *clientImpl) startReader() {
	defer constants.AutoRecover()()
	defer c.stop()

	msg := common.CreateMessage(c._client.opt.ByteOrder)
	defer common.DeleteMessage(msg)

	var bs = common.CreateByteBuffer(int(c._client.opt.MaxPacketSize))
	defer common.DeleteByteBuffer(bs)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			ret := func() (ret bool) {
				defer constants.AutoRecover()()
				err := c.readMessage(bs)
				if err != nil {
					logger.Log().Error("kcp client read error %v", err)
					return
				}
				err = msg.FromBytes(bs.Data)
				if err != nil {
					logger.Log().Error("kcp client msg.FromBytes %v", err)
					return
				}
				if c._client.opt.MessageCallback != nil {
					cmdid := int32(msg.Header >> 32)
					cmd := int32(msg.Header & 0xffffffff)
					if cmdid == gsinf.KcpControlCMD {
						if cmd > 0 {
							c._client.opt.MessageCallback.HandleControl(c, cmd, msg.ID)
						}
						return true
					}
					c._client.opt.MessageCallback.Handle(c, msg)
				}
				return true
			}()
			if !ret {
				return
			}
		}
	}
}

// StartWriter 写消息Goroutine， 用户将数据发送给客户端
func (c *clientImpl) startWriter() {
	defer constants.AutoRecover()()
	// logger.Log().Debug("kcp client [Writer Goroutine is running] id = %d", c.connId)
	// defer logger.Log().Debug("kcp client [conn Writer exit!] id = %d", c.connId)
	defer c.stop()

	//20秒检测一次,180秒视为连接关闭
	var interval_impl = time.Second * 20
	var _keeptimer = time.NewTimer(interval_impl)

	var bs = common.CreateByteBuffer(int(c._client.opt.MaxPacketSize))
	defer common.DeleteByteBuffer(bs)

	var ch = make(chan struct{})
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-_keeptimer.C:
			_keeptimer.Reset(interval_impl)
			//发送心跳, 0x7fffffff高位是控制消息
			c.SendControlMsg(0, ``)
		case msg, ok := <-c.writeBufferList:
			ret := func() (ret bool) {
				if ok {
					defer common.DeleteMessage(msg)
					if err := msg.ToBytes(bs); err != nil {
						logger.Log().Error(`kcp 消息打包错误 %+v`, err)
						return
					}
					deadline := time.Now().Add(c._client.opt.WriteTimeout)
					c.conn.SetWriteDeadline(deadline)
					_, err := c.conn.Write(bs.Data)
					if err != nil {
						logger.Log().Error(`kcp client Write error %+v`, err)
						return
					}
				} else {
					logger.Log().Debug("kcp client writeBufferList is Closed")
					return
				}
				return true
			}()
			if !ret {
				return
			}
		case <-ch:
		}
	}
}

func (c *clientImpl) sendRest() {
	defer constants.AutoRecover()()

	var bs = common.CreateByteBuffer(int(c._client.opt.MaxPacketSize))
	defer common.DeleteByteBuffer(bs)

	for {
		select {
		case msg, ok := <-c.writeBufferList:
			ret := func() (ret bool) {
				if ok {
					defer common.DeleteMessage(msg)
					if err := msg.ToBytes(bs); err != nil {
						logger.Log().Error(`kcp client 消息打包错误 %+v`, err)
						return
					}
					_, err := c.conn.Write(bs.Data)
					if err != nil {
						logger.Log().Error(`kcp Write error %+v`, err)
						return
					}
				} else {
					logger.Log().Debug("writeBufferList is Closed")
					return
				}
				return true
			}()
			if !ret {
				return
			}
		default:
			return
		}
	}
}
func (c *clientImpl) finalizer() {
	if c._client.opt.OnConnectorStop != nil {
		c._client.opt.OnConnectorStop(c)
	}

	logger.Log().Debug("kcp client Stop()...ConnID = %d", c.connId)

	// 关闭socket链接
	c.conn.Close()
	c.conn = nil

	c.running = 0

	c.writeBufferList = nil

	c._client.onImplError(c)
}
