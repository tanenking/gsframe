package zcommon

import (
	"sync"
)

type GroupMessage struct {
	GroupName string
	C         chan struct{}
	MsgID     string
	MsgData   []byte
}

const MaxGroupMsgCount uint16 = 8192

var GroupMsgSeq uint16
var GroupMsgSync sync.Mutex
var GroupMsgList = [MaxGroupMsgCount]*GroupMessage{}

func init() {
	GroupMsgSeq = 0
	GroupMsgList[GroupMsgSeq] = &GroupMessage{C: make(chan struct{})}
}

func SendGroupMsg(groupName string, msgID string, data []byte) {
	GroupMsgSync.Lock()
	gmsg := GroupMsgList[GroupMsgSeq]
	GroupMsgSeq++
	if GroupMsgSeq >= MaxGroupMsgCount {
		GroupMsgSeq = 0
	}
	GroupMsgList[GroupMsgSeq] = &GroupMessage{C: make(chan struct{})}
	GroupMsgSync.Unlock()

	gmsg.GroupName = groupName
	gmsg.MsgID = msgID
	gmsg.MsgData = data

	close(gmsg.C)
}
