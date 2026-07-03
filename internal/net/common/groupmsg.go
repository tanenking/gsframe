package common

type GroupMessage struct {
	GroupName string
	C         chan struct{}

	Header  int64
	MsgID   string
	MsgData []byte
}

const MaxGroupMsgCount uint16 = 8192
