package zws

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/logger"
	"github.com/tanenking/gsframe/internal/tcp/zcommon"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

// Connection 链接
type Connection struct {
	//当前Conn属于哪个Server
	TCPServer gsinf.IServer
	//当前连接的socket TCP套接字
	Conn *websocket.Conn
	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32
	//消息管理MsgID和对应处理方法的消息管理模块
	MsgHandler *zcommon.MsgHandle
	//告知该链接已经退出/停止的channel
	ctx    context.Context
	cancel context.CancelFunc
	//有缓冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan *zcommon.Message

	sync.RWMutex
	//链接属性
	//property map[string]interface{}
	property sync.Map
	////保护当前property的锁
	//propertyLock sync.Mutex
	//当前连接的关闭状态
	isClosed bool

	//读package
	rdpkg *zcommon.ReadPackage

	//心跳
	keepalive int
	//限流器
	limiter *rate.Limiter
	valid   bool
	real_ip string
	closing int32

	groupMsgSeq uint16
	groupMap    sync.Map
}

var connectionPool = sync.Pool{
	New: func() any {
		return &Connection{}
	},
}

// NewConnection 创建连接的方法
func NewConnection(server gsinf.IServer, conn *websocket.Conn, connID uint32, msgHandler *zcommon.MsgHandle, real_ip string) *Connection {
	//初始化Conn属性
	c := &Connection{
		TCPServer:   server,
		Conn:        conn,
		ConnID:      connID,
		isClosed:    false,
		MsgHandler:  msgHandler,
		msgBuffChan: make(chan *zcommon.Message, zcommon.GlobalObject.MaxMsgChanLen),
		property:    sync.Map{},
		rdpkg:       zcommon.NewReadPackage(),
		limiter:     rate.NewLimiter(zcommon.Limiter_limit, zcommon.Limiter_bucket),
		valid:       false,
		groupMsgSeq: zcommon.GroupMsgSeq,
		groupMap:    sync.Map{},
	}
	c.real_ip = real_ip

	tcpConn, ok := conn.UnderlyingConn().(*net.TCPConn)
	if ok {
		// 开启 NoDelay: 关闭Nagle, 小包立即发送
		_ = tcpConn.SetNoDelay(zcommon.GlobalObject.NoDelay)

		// 设置读写缓冲区
		_ = tcpConn.SetReadBuffer(int(zcommon.GlobalObject.TcpReadWriteBufferSize))
		_ = tcpConn.SetWriteBuffer(int(zcommon.GlobalObject.TcpReadWriteBufferSize))

		// 开启 TCP Keepalive
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	return c
}

func (c *Connection) InGroup(groupName string) bool {
	_, ok := c.groupMap.Load(groupName)
	return ok
}
func (c *Connection) AddGroup(groupName string) {
	c.groupMap.Store(groupName, struct{}{})
}
func (c *Connection) DeleteGroup(groupName string) {
	c.groupMap.Delete(groupName)
}
func (c *Connection) GetGroupList() []string {
	list := []string{}
	c.groupMap.Range(func(key, value any) bool {
		list = append(list, key.(string))
		return true
	})
	return list
}

// StartWriter 写消息Goroutine， 用户将数据发送给客户端
func (c *Connection) StartWriter() {
	defer constants.AutoRecover()()
	logger.Log().Debug("[Writer Goroutine is running]")
	defer logger.Log().Debug("%s [conn Writer exit!]", c.ClientIP())
	defer c.Stop()

	//20秒检测一次,180秒视为连接关闭,20秒无player属性视为非法
	interval_impl := time.Second * 20
	_keeptimer := time.NewTimer(interval_impl)
	c.keepalive = 0

	// timer := time.NewTimer(time.Millisecond * 20)
	gmsg := zcommon.GroupMsgList[c.groupMsgSeq]
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-_keeptimer.C:
			c.keepalive++
			if c.keepalive >= 9 {
				logger.Log().Error("tcp心跳超时")
				return
			} else if !c.IsValid() {
				logger.Log().Error("tcp连接20秒内都没有绑定player")
				return
			}
			_keeptimer.Reset(interval_impl)
			deadline := time.Now().Add(zcommon.GlobalObject.WriteTimeout)
			if err := c.Conn.SetWriteDeadline(deadline); err != nil {
				logger.Log().Error(`SetWriteDeadline error %+v`, err)
				return
			}
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				logger.Log().Error(`websocket.PingMessage error %+v`, err)
				return
			}
		case data, ok := <-c.msgBuffChan:
			ret := func() (ret bool) {
				if ok {
					defer func() {
						zcommon.MessagePoop.Put(data)
					}()
					msg, err := zcommon.Pack(data)
					if err != nil {
						logger.Log().Error(`消息打包错误 %+v`, err)
						return
					}
					err = c.WriteMessage(msg)
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
				if c.SendBuffMsg(context.Background(), gmsg.MsgID, gmsg.MsgData) != nil {
					return
				}
			}
			c.groupMsgSeq++
			if c.groupMsgSeq >= zcommon.MaxGroupMsgCount {
				c.groupMsgSeq = 0
			}
			gmsg = zcommon.GroupMsgList[c.groupMsgSeq]
			// case <-timer.C:
			// 	timer.Reset(time.Millisecond * 20)
		}
	}
}

// StartReader 读消息Goroutine，用于从客户端中读取数据
func (c *Connection) StartReader() {
	defer constants.AutoRecover()()
	logger.Log().Debug("[Reader Goroutine is running]")
	defer logger.Log().Debug("%s [conn Reader exit!]", c.ClientIP())
	defer c.Stop()

	// 创建拆包解包的对象
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			ret := func() (ret bool) {
				defer constants.AutoRecover()()
				wsmsg := zcommon.MessagePoop.Get().(*zcommon.Message)
				defer func() {
					zcommon.MessagePoop.Put(wsmsg)
				}()
				// t, b, err := c.Conn.ReadMessage()
				t, err := zcommon.ReadWSMessage(c.Conn, wsmsg)
				if err != nil {
					logger.Log().Error("ReadWSMessage %v", err)
					return
				}
				c.keepalive = 0
				c.Conn.SetReadDeadline(time.Now().Add(zcommon.GlobalObject.ReadTimeout))
				switch t {
				case websocket.TextMessage:
					logger.Log().Debug("TextMessage -> %s", string(wsmsg.Data))
				case websocket.PingMessage:
					logger.Log().Debug("PingMessage -> %s", string(wsmsg.Data))
				case websocket.PongMessage:
					// logger.Log().Debug("PongMessage -> %s", string(wsmsg.Data))
				case websocket.CloseMessage:
					logger.Log().Debug("CloseMessage -> %s", string(wsmsg.Data))
					return
				case websocket.BinaryMessage:
					msg := zcommon.MessagePoop.Get().(*zcommon.Message)
					defer func() {
						zcommon.MessagePoop.Put(msg)
					}()
					err := zcommon.UnpackFromBytes(wsmsg.Data, wsmsg.DataLen, msg)
					if err != nil {
						logger.Log().Error("UnpackFromBytes %v", err)
						return
					}
					req := zcommon.RequestPool.Get().(*zcommon.Request)
					defer func() {
						zcommon.RequestPool.Put(req)
					}()

					req.Conn = c
					req.Msg = msg
					c.MsgHandler.DoMsgHandler(req)
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

func (c *Connection) sendRest() {
	defer constants.AutoRecover()()
	defer logger.Log().Debug("%s [sendRest!]", c.ClientIP())
	for {
		select {
		case data, ok := <-c.msgBuffChan:
			if !ok || data == nil {
				return
			}
			ret := func() (ret bool) {
				defer func() {
					zcommon.MessagePoop.Put(data)
				}()
				msg, err := zcommon.Pack(data)
				if err != nil {
					logger.Log().Error(`消息打包错误 %+v`, err)
					return
				}
				err = c.WriteMessage(msg)
				if err != nil {
					logger.Log().Error(`WriteMessage error %+v`, err)
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

func (c *Connection) WriteMessage(msg []byte) error {
	deadline := time.Now().Add(zcommon.GlobalObject.WriteTimeout)
	if err := c.Conn.SetWriteDeadline(deadline); err != nil {
		return err
	}
	err := c.Conn.WriteMessage(websocket.BinaryMessage, msg)
	if err == nil {
		return err
	}

	var closeErr *websocket.CloseError
	if errors.As(err, &closeErr) || websocket.IsUnexpectedCloseError(err) {
		return errors.New("websocket connection closed: " + err.Error())
	}
	// 超时错误，可短暂重试一次（可选）
	if errors.Is(err, context.DeadlineExceeded) {
		// 重试一次
		time.Sleep(100 * time.Millisecond)
		return c.Conn.WriteMessage(websocket.BinaryMessage, msg)
	}
	return nil
}

// Start 启动连接，让当前连接开始工作
func (c *Connection) Start() {
	c.ctx, c.cancel = context.WithCancel(context.Background())
	//1 开启用户从客户端读取数据流程的Goroutine
	constants.Go(func() {
		c.StartReader()
	})
	//2 开启用于写回客户端数据流程的Goroutine
	constants.Go(func() {
		c.StartWriter()
	})

	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	c.TCPServer.CallOnConnStart(c)

	<-c.ctx.Done()

	//关闭该链接全部管道
	close(c.msgBuffChan)

	//关闭连接,将还未发送完的数据发完
	c.sendRest()

	c.finalizer()
}

// Stop 停止连接，结束当前连接状态M
func (c *Connection) Stop() {
	if !atomic.CompareAndSwapInt32(&c.closing, 0, 1) {
		return
	}
	constants.Go(func() {
		c.cancel()
	})
}

// GetConnID 获取当前连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// RemoteAddr 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}
func (c *Connection) ClientIP() string {
	return c.real_ip
}
func (c *Connection) ClientAddress() string { //ip:port
	addr := c.Conn.RemoteAddr().String()
	addrs := strings.Split(addr, ":")
	port := "0"
	if len(addrs) >= 2 {
		port = addrs[1]
	}
	return fmt.Sprintf(`%s:%s`, c.real_ip, port)
}

// SendBuffMsg  发生BuffMsg
func (c *Connection) SendBuffMsg(ctx context.Context, msgID string, data []byte) error {
	defer constants.AutoRecover()()
	c.RLock()
	defer c.RUnlock()
	if c.isClosed {
		return errors.New("Connection closed when send buff msg")
	}

	msg := zcommon.MessagePoop.Get().(*zcommon.Message)
	msg.ID = msgID
	msg.DataLen = uint32(len(data))
	msg.Data = msg.Data[:msg.DataLen]
	copy(msg.Data, data)
	// msg.Data = append(msg.Data, data...)
	msg.RequestSeq = constants.ParseSeq(ctx)
	//写回客户端
	c.msgBuffChan <- msg //zcommon.NewMsgPackage(msgID, data, constants.ParseSeq(ctx))

	return nil
}

// SetProperty 设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	if value != nil {
		c.property.Store(key, value)
	} else {
		c.property.Delete(key)
	}
}

// GetProperty 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	if value, ok := c.property.Load(key); ok && value != nil {
		return value, nil
	}

	return nil, errors.New("no property found")
}

// RemoveProperty 移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.property.Delete(key)
}

// 返回ctx，用于用户自定义的go程获取连接退出状态
func (c *Connection) Context() context.Context {
	return c.ctx
}

func (c *Connection) finalizer() {
	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.TCPServer.CallOnConnStop(c)

	c.Lock()
	defer c.Unlock()

	//如果当前链接已经关闭
	if c.isClosed {
		return
	}

	logger.Log().Debug("Conn Stop()...ConnID = %d", c.ConnID)

	// 关闭socket链接
	_ = c.Conn.Close()

	//将链接从连接管理器中删除
	c.TCPServer.GetConnMgr().Remove(c)

	c.msgBuffChan = nil
	//设置标志位
	c.isClosed = true
}
func (c *Connection) GetLimiterToken() error {
	ctx, cancel := context.WithTimeout(context.Background(), zcommon.Limiter_Timeout)
	defer cancel()
	return c.limiter.Wait(ctx)
}
func (c *Connection) IsValid() bool { //是否有效连接
	if c.valid {
		return true
	}
	if p, err := c.GetProperty("player"); err == nil && p != nil {
		c.valid = true
	}
	return c.valid
}
func (c *Connection) IsClosing() bool {
	return atomic.LoadInt32(&c.closing) > 0
}
