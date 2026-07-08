package kcpx

import (
	"encoding/binary"
	"time"

	"github.com/tanenking/gsframe/gsinf"
)

/*
定义一个全局的对象
*/
var config *gsinf.KcpServerConfig

/*
提供init方法，默认加载
*/
func init() {
	//初始化GlobalObject变量，设置一些默认值
	config = &gsinf.KcpServerConfig{
		Port:                   0,
		GoReadWriteBufferSize:  4096,
		TcpReadWriteBufferSize: 256 * 1024, //256k
		MaxPacketSize:          4096,
		WriteMessageBufferLen:  4096,
		ReadTimeout:            time.Minute,
		WriteTimeout:           time.Second * 10,
		HeartTimeoutSec:        120,
		NoDelay:                true,
		StreamMode:             false,
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
		config.WriteMessageBufferLen = 4096
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

func validateClientConfig(opt *gsinf.KcpClientConfig) {
	if opt.PoolSize <= 0 {
		opt.PoolSize = 30
	}
	if opt.GoReadWriteBufferSize <= 0 {
		opt.GoReadWriteBufferSize = 4096
	} else if opt.GoReadWriteBufferSize > 65535 {
		opt.GoReadWriteBufferSize = 65535
	}
	if opt.TcpReadWriteBufferSize <= 0 {
		opt.TcpReadWriteBufferSize = 256 * 1024
	}
	if opt.MaxPacketSize <= 0 {
		opt.MaxPacketSize = 4096
	} else if opt.MaxPacketSize > 65535 {
		opt.MaxPacketSize = 65535
	}
	if opt.WriteMessageBufferLen <= 0 {
		opt.WriteMessageBufferLen = 4096
	}
	if opt.ReadTimeout <= 0 {
		opt.ReadTimeout = time.Minute
	}
	if opt.WriteTimeout <= 0 {
		opt.WriteTimeout = time.Second * 10
	}
	if opt.ByteOrder == nil {
		opt.ByteOrder = binary.LittleEndian
	}
}
