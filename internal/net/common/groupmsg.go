package common

import (
	"sync"
)

type GroupMessage struct {
	GroupName string
	C         chan struct{}

	Header  int64
	MsgID   string
	MsgData []byte
}

const MaxGroupMsgCount uint16 = 8192

var GroupMsgSeq uint16
var GroupMsgSync sync.Mutex
var GroupMsgList = [MaxGroupMsgCount]*GroupMessage{}

func init() {
	GroupMsgSeq = 0
	GroupMsgList[GroupMsgSeq] = &GroupMessage{C: make(chan struct{})}
}

func SendGroup(groupName string, header int64, msgID string, data []byte) {
	GroupMsgSync.Lock()
	gmsg := GroupMsgList[GroupMsgSeq]
	GroupMsgSeq++
	if GroupMsgSeq >= MaxGroupMsgCount {
		GroupMsgSeq = 0
	}
	GroupMsgList[GroupMsgSeq] = &GroupMessage{C: make(chan struct{})}
	GroupMsgSync.Unlock()

	gmsg.GroupName = groupName
	gmsg.Header = header
	gmsg.MsgID = msgID
	if cap(gmsg.MsgData) < len(data) {
		gmsg.MsgData = make([]byte, 0, len(data))
	}
	gmsg.MsgData = gmsg.MsgData[:len(data)]
	copy(gmsg.MsgData, data)

	close(gmsg.C)
}
