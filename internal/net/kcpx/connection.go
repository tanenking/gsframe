package kcpx

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/logger"
	"github.com/tanenking/gsframe/internal/net/common"
	"github.com/xtaci/kcp-go/v5"
)

var connectionPool = sync.Pool{
	New: func() any {
		return newConnection()
	},
}

type connection struct {
	_server       *server
	ctx           context.Context
	cancel        context.CancelFunc
	conn          *kcp.UDPSession
	connId        int32
	clientIp      string
	property      sync.Map
	closed        int32
	lastHeartTime int64

	writeBufferList chan *common.Message

	// groupMsgSeq uint16
	// groupMap    sync.Map
}

func newConnection() *connection {
	return &connection{}
}

func (c *connection) init(_server *server, conn *kcp.UDPSession, connId int32) bool {
	c._server = _server

	c.ctx, c.cancel = context.WithCancel(context.Background())

	c.conn = conn
	c.closed = 0
	c.connId = connId
	c.property = sync.Map{}
	c.lastHeartTime = time.Now().Unix()

	c.writeBufferList = make(chan *common.Message, config.WriteMessageBufferLen)

	// c.groupMsgSeq = c._server.groupMsgSeq
	// c.groupMap = sync.Map{}

	c.conn.SetReadDeadline(time.Time{})
	c.conn.SetWriteDeadline(time.Now().Add(config.WriteTimeout))

	c.conn.SetRateLimit(0)
	c.conn.SetStreamMode(config.StreamMode)
	c.conn.SetNoDelay(1, 10, 2, 1)
	c.conn.SetACKNoDelay(true)
	c.conn.SetWindowSize(1024, 1024)
	c.conn.SetMtu(1472)
	c.conn.SetWriteDelay(true)

	c.conn.SetReadBuffer(int(config.TcpReadWriteBufferSize))
	c.conn.SetWriteBuffer(int(config.TcpReadWriteBufferSize))

	return true
}

func (c *connection) ClientIP() string {
	return c.clientIp
}

// Stop 停止连接，结束当前连接状态M
func (c *connection) Stop() {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return
	}
	constants.Go(func() {
		c.cancel()
	})
}

func (c *connection) GetConnID() int32 {
	return c.connId
}

func (c *connection) IsValid() bool {
	return atomic.LoadInt32(&c.closed) == 0 && c.conn != nil
}

// func (c *connection) InGroup(groupName string) bool {
// 	_, ok := c.groupMap.Load(groupName)
// 	return ok
// }
// func (c *connection) AddGroup(groupName string) {
// 	c.groupMap.Store(groupName, struct{}{})
// }
// func (c *connection) DeleteGroup(groupName string) {
// 	c.groupMap.Delete(groupName)
// }
// func (c *connection) GetGroupList() []string {
// 	list := []string{}
// 	c.groupMap.Range(func(key, value any) bool {
// 		list = append(list, key.(string))
// 		return true
// 	})
// 	return list
// }

// SetProperty 设置链接属性
func (c *connection) SetProperty(key string, value interface{}) {
	if value != nil {
		c.property.Store(key, value)
	} else {
		c.property.Delete(key)
	}
}

// GetProperty 获取链接属性
func (c *connection) GetProperty(key string) (interface{}, error) {
	if value, ok := c.property.Load(key); ok && value != nil {
		return value, nil
	}

	return nil, errors.New("no property found")
}

// RemoveProperty 移除链接属性
func (c *connection) RemoveProperty(key string) {
	c.property.Delete(key)
}

func (c *connection) Send(header int64, msgID string, data []byte) error {
	defer constants.AutoRecover()()
	if atomic.LoadInt32(&c.closed) > 0 {
		return fmt.Errorf("kcp Connection %d closed when send buff msg", c.connId)
	}

	var msg = common.CreateMessage(config.ByteOrder)
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

func (c *connection) start() {
	if config.OnConnectionCreate != nil {
		config.OnConnectionCreate(c)
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

func (c *connection) readMessage(bs *common.ByteBuffer) (err error) {
	if c.conn == nil {
		return errors.New(`kcp readMessage conn is nil`)
	}
	bs.Data = bs.Data[:0]
	var recvtotal = 0
	var recv = make([]byte, cap(bs.Data))
	for {
		recv = recv[:cap(recv)]
		var recvcount int
		recvcount, err = c.conn.Read(recv)
		if err != nil {
			logger.Log().Error("kcp read error %v", err)
			return
		}
		if recvcount == 0 {
			return
		}
		recv = recv[:recvcount]
		if cap(bs.Data) < (recvtotal + recvcount) {
			//扩容
			newBuf := make([]byte, len(bs.Data), cap(bs.Data)*2)
			copy(newBuf, bs.Data)
			bs.Data = newBuf
		}
		bs.Data = bs.Data[:(recvtotal + recvcount)]
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
func (c *connection) startReader() {
	defer constants.AutoRecover()()
	defer c.Stop()

	msg := common.CreateMessage(config.ByteOrder)
	defer common.DeleteMessage(msg)

	var bs = common.CreateByteBuffer(int(config.MaxPacketSize))
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
					logger.Log().Error("kcp read error %v", err)
					return
				}
				c.lastHeartTime = time.Now().Unix()
				err = msg.FromBytes(bs.Data)
				if err != nil {
					logger.Log().Error("kcp msg.FromBytes %v", err)
					return
				}
				if config.MessageCallback != nil {
					cmdid := int32(msg.Header >> 32)
					cmd := int32(msg.Header & 0xffffffff)
					if cmdid == gsinf.KcpControlCMD {
						if cmd > 0 {
							config.MessageCallback.HandleControl(c, cmd, msg.ID)
						}
						return true
					}
					if !config.MessageCallback.PreHandle(c, msg) {
						return true
					}
					config.MessageCallback.Handle(c, msg)
					config.MessageCallback.PostHandle(c, msg)
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
func (c *connection) startWriter() {
	defer constants.AutoRecover()()
	defer c.Stop()

	//20秒检测一次,180秒视为连接关闭
	var interval_impl = time.Second * 20
	var _keeptimer = time.NewTimer(interval_impl)

	var bs = common.CreateByteBuffer(int(config.MaxPacketSize))
	defer common.DeleteByteBuffer(bs)

	var ch = make(chan struct{})
	// gmsg := c._server.groupMsgList[c.groupMsgSeq]
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-_keeptimer.C:
			if time.Now().Unix()-c.lastHeartTime >= config.HeartTimeoutSec {
				logger.Log().Error("udp 心跳超时")
				return
			}
			_keeptimer.Reset(interval_impl)
		case msg, ok := <-c.writeBufferList:
			ret := func() (ret bool) {
				if ok {
					defer common.DeleteMessage(msg)
					if err := msg.ToBytes(bs); err != nil {
						logger.Log().Error(`kcp 消息打包错误 %+v`, err)
						return
					}
					deadline := time.Now().Add(config.WriteTimeout)
					c.conn.SetWriteDeadline(deadline)
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
		// case <-gmsg.C:
		// 	if c.InGroup(gmsg.GroupName) {
		// 		if c.Send(gmsg.Header, gmsg.MsgID, gmsg.MsgData) != nil {
		// 			return
		// 		}
		// 	}
		// 	c.groupMsgSeq++
		// 	if c.groupMsgSeq >= common.MaxGroupMsgCount {
		// 		c.groupMsgSeq = 0
		// 	}
		// 	gmsg = c._server.groupMsgList[c.groupMsgSeq]
		case <-ch:
		}
	}
}

func (c *connection) sendRest() {
	defer constants.AutoRecover()()
	defer logger.Log().Debug("kcp [sendRest!] id = %d", c.connId)

	var bs = common.CreateByteBuffer(int(config.MaxPacketSize))
	defer common.DeleteByteBuffer(bs)

	for {
		select {
		case msg, ok := <-c.writeBufferList:
			ret := func() (ret bool) {
				if ok {
					defer common.DeleteMessage(msg)
					if err := msg.ToBytes(bs); err != nil {
						logger.Log().Error(`kcp 消息打包错误 %+v`, err)
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

func (c *connection) finalizer() {
	//如果当前链接已经关闭
	if atomic.LoadInt32(&c.closed) > 0 {
		return
	}

	if config.OnConnectionStop != nil {
		config.OnConnectionStop(c)
	}

	logger.Log().Debug("kcp Conn Stop()...ConnID = %d", c.connId)

	// 关闭socket链接
	c.conn.Close()
	c.conn = nil

	//将链接从连接管理器中删除
	c._server.freeConnection(c)

	c.writeBufferList = nil
}
