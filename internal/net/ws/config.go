package ws

import (
	"encoding/binary"
	"time"

	"github.com/tanenking/gsframe/gsinf"
)

/*
定义一个全局的对象
*/
var config *gsinf.WebSocketServerConfig

/*
提供init方法，默认加载
*/
func init() {
	//初始化GlobalObject变量，设置一些默认值
	config = &gsinf.WebSocketServerConfig{
		Port:                   0,
		MaxConn:                5000,
		GoReadWriteBufferSize:  4096,
		TcpReadWriteBufferSize: 256 * 1024, //256k
		MaxPacketSize:          65536,
		WriteMessageBufferLen:  128,
		ReadTimeout:            time.Minute,
		WriteTimeout:           time.Second * 10,
		HeartTimeoutSec:        120,
		NoDelay:                true,
		LimiterLimit:           0, //rate.Every(time.Millisecond * 100),
		LimiterTimeout:         time.Second,
		LimiterBucketCount:     10,
		ByteOrder:              binary.LittleEndian,
	}
}

func Validate() {
	if config.MaxConn < 3000 {
		config.MaxConn = 3000
	} else if config.MaxConn > 50000 {
		config.MaxConn = 50000
	}
	if config.MaxPacketSize < 8192 {
		config.MaxPacketSize = 8192
	} else if config.MaxPacketSize > 65535 {
		config.MaxPacketSize = 65535
	}
}
