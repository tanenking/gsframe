package zcommon

import (
	"time"

	"github.com/tanenking/gsframe/gsinf"
)

/*
定义一个全局的对象
*/
var GlobalObject *gsinf.TcpGlobalConfig_t

/*
提供init方法，默认加载
*/
func init() {
	//初始化GlobalObject变量，设置一些默认值
	GlobalObject = &gsinf.TcpGlobalConfig_t{
		TCPPort:       8999,
		Host:          "0.0.0.0",
		MaxConn:       12000,
		MaxPacketSize: 8192,
		MaxMsgChanLen: 128,
		ReadTimeout:   time.Minute,
		WriteTimeout:  time.Second * 10,
	}
}

func Validate() {
	if GlobalObject.MaxConn < 3000 {
		GlobalObject.MaxConn = 3000
	} else if GlobalObject.MaxConn > 50000 {
		GlobalObject.MaxConn = 50000
	}
	if GlobalObject.MaxPacketSize < 8192 {
		GlobalObject.MaxPacketSize = 8192
	} else if GlobalObject.MaxPacketSize > 65535 {
		GlobalObject.MaxPacketSize = 65535
	}
}
