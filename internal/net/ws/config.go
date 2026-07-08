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

func validateConfig() {
	if config.GoReadWriteBufferSize <= 0 {
		config.GoReadWriteBufferSize = 4096
	} else if config.GoReadWriteBufferSize > 65535 {
		config.GoReadWriteBufferSize = 65535
	}
	if config.TcpReadWriteBufferSize <= 0 {
		config.TcpReadWriteBufferSize = 256 * 1024
	}
	if config.MaxPacketSize <= 0 {
		config.MaxPacketSize = 4096
	} else if config.MaxPacketSize > 65535 {
		config.MaxPacketSize = 65535
	}
	if config.WriteMessageBufferLen <= 0 {
		config.WriteMessageBufferLen = 128
	}
	if config.ReadTimeout <= 0 {
		config.ReadTimeout = time.Minute
	}
	if config.WriteTimeout <= 0 {
		config.WriteTimeout = time.Second * 10
	}
	if config.HeartTimeoutSec <= 0 {
		config.HeartTimeoutSec = 120
	}
	if config.ByteOrder == nil {
		config.ByteOrder = binary.LittleEndian
	}
}
