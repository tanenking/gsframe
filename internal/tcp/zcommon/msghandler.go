package zcommon

import (
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/logx"
)

// MsgHandle -
type MsgHandle struct {
	GlobalApi      gsinf.IRouter
	Apis           map[string]gsinf.IRouter //存放每个MsgID 所对应的处理方法的map属性
	WorkerPoolSize uint32                   //业务工作Worker池的数量
	TaskQueue      []chan gsinf.IRequest    //Worker负责取任务的消息队列
}

// NewMsgHandle 创建MsgHandle
func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis: make(map[string]gsinf.IRouter),
	}
}

func (mh *MsgHandle) DoMsgHandler(request gsinf.IRequest) {
	var handler gsinf.IRouter
	msgid := request.GetMsgID()
	if len(msgid) <= 0 {
		handler = mh.GlobalApi
	} else {
		var ok bool
		handler, ok = mh.Apis[request.GetMsgID()]
		if !ok {
			//没有注册的消息,统一通过global处理
			handler = mh.GlobalApi
		}
	}
	if handler == nil {
		logx.DebugF("api msgID = %s is not FOUND!", request.GetMsgID())
		return
	}

	//执行对应处理方法
	if !handler.PreHandle(request) {
		return
	}
	handler.Handle(request)
	handler.PostHandle(request)
}

// RegisterRouter 为消息添加具体的处理逻辑
func (mh *MsgHandle) RegisterRouter(msgID string, router gsinf.IRouter) {
	//1 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[msgID]; ok {
		panic("repeated api , msgID = " + msgID)
	}
	//2 添加msg与api的绑定关系
	mh.Apis[msgID] = router
	logx.DebugF("Add api msgID = %s", msgID)
}

// 路由功能: 没有指定的消息,都通过这个处理
func (mh *MsgHandle) RegisterGlobalRouter(router gsinf.IRouter) {
	mh.GlobalApi = router
}
