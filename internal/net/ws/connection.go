package ws

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/logger"
	"github.com/tanenking/gsframe/internal/net/common"
	"golang.org/x/time/rate"
)

type connection struct {
	_server       *server
	ctx           context.Context
	cancel        context.CancelFunc
	conn          *websocket.Conn
	connId        int32
	clientIp      string
	property      sync.Map
	limiter       *rate.Limiter
	closed        int32
	lastHeartTime int64

	writeBufferList chan *common.Message

	groupMsgSeq uint16
	groupMap    sync.Map
}

func newConnection() *connection {
	return &connection{}
}
func (r *connection) init(_server *server, conn *websocket.Conn, connId int32) bool {
	r._server = _server

	r.ctx, r.cancel = context.WithCancel(context.Background())

	r.conn = conn
	r.closed = 0
	r.connId = connId
	r.property = sync.Map{}
	r.lastHeartTime = time.Now().Unix()

	r.writeBufferList = make(chan *common.Message, config.WriteMessageBufferLen)

	r.groupMsgSeq = common.GroupMsgSeq
	r.groupMap = sync.Map{}

	if config.LimiterLimit > 0 {
		r.limiter = rate.NewLimiter(config.LimiterLimit, int(config.LimiterBucketCount))
	}

	_ = conn.SetReadDeadline(time.Now().Add(config.ReadTimeout))
	_ = conn.SetWriteDeadline(time.Now().Add(config.WriteTimeout))

	// 设置Pong回调 (刷新读超时，保持空闲连接存活)
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(config.ReadTimeout))
		return nil
	})

	tcpConn, ok := conn.UnderlyingConn().(*net.TCPConn)
	if ok {
		// 开启 NoDelay: 关闭Nagle, 小包立即发送
		_ = tcpConn.SetNoDelay(config.NoDelay)

		// 设置读写缓冲区
		_ = tcpConn.SetReadBuffer(int(config.TcpReadWriteBufferSize))
		_ = tcpConn.SetWriteBuffer(int(config.TcpReadWriteBufferSize))

		// 开启 TCP Keepalive
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

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
func (c *connection) InGroup(groupName string) bool {
	_, ok := c.groupMap.Load(groupName)
	return ok
}
func (c *connection) AddGroup(groupName string) {
	c.groupMap.Store(groupName, struct{}{})
}
func (c *connection) DeleteGroup(groupName string) {
	c.groupMap.Delete(groupName)
}
func (c *connection) GetGroupList() []string {
	list := []string{}
	c.groupMap.Range(func(key, value any) bool {
		list = append(list, key.(string))
		return true
	})
	return list
}
func (c *connection) WaitLimiterToken() error {
	if c.limiter == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.LimiterTimeout)
	defer cancel()
	return c.limiter.Wait(ctx)
}

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
	if c.closed > 0 {
		return fmt.Errorf("Connection %d closed when send buff msg", c.connId)
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

// //////////////////////////////////////////////////////////////////////////////
// Start 启动连接，让当前连接开始工作
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

func (c *connection) readMessage(conn *websocket.Conn, bs []byte) (int, error) {
	if conn == nil {
		return 0, errors.New(`readMessage conn is nil`)
	}
	bs = bs[:0]
	msgType, reader, err := conn.NextReader()
	if err != nil {
		return 0, err
	}

	for {
		if len(bs) >= cap(bs)-1024 {
			newBuf := make([]byte, len(bs), cap(bs)*2)
			copy(newBuf, bs)
			bs = newBuf
		}

		n, err := reader.Read(bs[len(bs):cap(bs)])
		// if n > 0 {
		// 	bs = bs[:len(bs)+n]
		// }
		if n <= 0 {

		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}

	return msgType, nil
}

// StartReader 读消息Goroutine，用于从客户端中读取数据
func (c *connection) startReader() {
	defer constants.AutoRecover()()
	logger.Log().Debug("[Reader Goroutine is running] id = %d", c.connId)
	defer logger.Log().Debug("[conn Reader exit!] id = %d", c.connId)
	defer c.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			ret := func() (ret bool) {
				defer constants.AutoRecover()()
				var bs = common.CreateByteBuffer(4096)
				defer func() {
					common.DeleteByteBuffer(bs)
				}()
				t, err := c.readMessage(c.conn, bs)
				if err != nil {
					logger.Log().Error("readWSMessage %v", err)
					return
				}
				c.lastHeartTime = time.Now().Unix()
				c.conn.SetReadDeadline(time.Now().Add(config.ReadTimeout))
				switch t {
				case websocket.TextMessage:
					logger.Log().Debug("TextMessage -> %s", string(bs))
				case websocket.PingMessage:
					logger.Log().Debug("PingMessage -> %s", string(bs))
				case websocket.PongMessage:
					// logger.Log().Debug("PongMessage -> %s", string(bs))
				case websocket.CloseMessage:
					logger.Log().Debug("CloseMessage -> %s", string(bs))
					return
				case websocket.BinaryMessage:
					msg := common.CreateMessage(config.ByteOrder)
					defer func() {
						common.DeleteMessage(msg)
					}()
					err := msg.FromBytes(bs)
					if err != nil {
						logger.Log().Error("msg.FromBytes %v", err)
						return
					}
					if config.MessageCallback != nil {
						if !config.MessageCallback.PreHandle(c, msg) {
							return true
						}
						config.MessageCallback.Handle(c, msg)
						config.MessageCallback.PostHandle(c, msg)
					}
				default:
					logger.Log().Error("非法消息类型 -> %d", t)
					return
				}
				return true
			}()
			if !ret {
				return
			}
			time.Sleep(time.Millisecond * 20)
		}
	}
}

func (c *connection) writeMessage(data []byte) error {
	deadline := time.Now().Add(config.WriteTimeout)
	if err := c.conn.SetWriteDeadline(deadline); err != nil {
		return err
	}
	err := c.conn.WriteMessage(websocket.BinaryMessage, data)
	if err == nil {
		return err
	}

	if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
		return errors.New("websocket connection closed: " + err.Error())
	}
	// 超时错误，可短暂重试一次（可选）
	if errors.Is(err, context.DeadlineExceeded) {
		// 重试一次
		time.Sleep(100 * time.Millisecond)
		return c.conn.WriteMessage(websocket.BinaryMessage, data)
	}
	return nil
}

// StartWriter 写消息Goroutine， 用户将数据发送给客户端
func (c *connection) startWriter() {
	defer constants.AutoRecover()()
	logger.Log().Debug("[Writer Goroutine is running] id = %d", c.connId)
	defer logger.Log().Debug("[conn Writer exit!] id = %d", c.connId)
	defer c.Stop()

	//20秒检测一次,180秒视为连接关闭,20秒无player属性视为非法
	var interval_impl = time.Second * 20
	var _keeptimer = time.NewTimer(interval_impl)

	gmsg := common.GroupMsgList[c.groupMsgSeq]
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-_keeptimer.C:
			if time.Now().Unix()-c.lastHeartTime >= config.HeartTimeoutSec {
				logger.Log().Error("tcp心跳超时")
				return
			}
			_keeptimer.Reset(interval_impl)
			deadline := time.Now().Add(config.WriteTimeout)
			if err := c.conn.SetWriteDeadline(deadline); err != nil {
				logger.Log().Error(`SetWriteDeadline error %+v`, err)
				return
			}
			err := c.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				logger.Log().Error(`websocket.PingMessage error %+v`, err)
				return
			}
		case msg, ok := <-c.writeBufferList:
			ret := func() (ret bool) {
				if ok {
					defer func() {
						common.DeleteMessage(msg)
					}()
					var bs = common.CreateByteBuffer(4096)
					defer func() {
						common.DeleteByteBuffer(bs)
					}()
					if err := msg.ToBytes(bs); err != nil {
						logger.Log().Error(`消息打包错误 %+v`, err)
						return
					}
					var err = c.writeMessage(bs)
					if err != nil {
						logger.Log().Error(`WriteMessage error %+v`, err)
						return
					}
				} else {
					logger.Log().Debug("msgBuffChan is Closed")
					return
				}
				return true
			}()
			if !ret {
				return
			}
		case <-gmsg.C:
			if c.InGroup(gmsg.GroupName) {
				if c.Send(gmsg.Header, gmsg.MsgID, gmsg.MsgData) != nil {
					return
				}
			}
			c.groupMsgSeq++
			if c.groupMsgSeq >= common.MaxGroupMsgCount {
				c.groupMsgSeq = 0
			}
			gmsg = common.GroupMsgList[c.groupMsgSeq]
		}
	}
}

func (c *connection) sendRest() {
	defer constants.AutoRecover()()
	defer logger.Log().Debug("[sendRest!] id = %d", c.connId)
	for {
		select {
		case msg, ok := <-c.writeBufferList:
			ret := func() (ret bool) {
				if ok {
					defer func() {
						common.DeleteMessage(msg)
					}()
					var bs = common.CreateByteBuffer(4096)
					defer func() {
						common.DeleteByteBuffer(bs)
					}()
					if err := msg.ToBytes(bs); err != nil {
						logger.Log().Error(`消息打包错误 %+v`, err)
						return
					}
					var err = c.writeMessage(bs)
					if err != nil {
						logger.Log().Error(`WriteMessage error %+v`, err)
						return
					}
				} else {
					logger.Log().Debug("msgBuffChan is Closed")
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
	if config.OnConnectionStop != nil {
		config.OnConnectionStop(c)
	}

	//如果当前链接已经关闭
	if atomic.LoadInt32(&c.closed) > 0 {
		return
	}

	logger.Log().Debug("Conn Stop()...ConnID = %d", c.connId)

	// 关闭socket链接
	_ = c.conn.Close()
	c.conn = nil

	//将链接从连接管理器中删除
	c._server.connections[c.connId] = nil
	c._server.freeConnection(c)

	c.writeBufferList = nil
}
