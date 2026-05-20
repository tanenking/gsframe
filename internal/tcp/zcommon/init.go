package zcommon

import (
	"encoding/binary"
	"errors"
	"io"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tanenking/gsframe/internal/logx"

	"golang.org/x/time/rate"
)

var (
	Limiter_limit   rate.Limit       = rate.Every(time.Millisecond * 100)
	Limiter_Timeout time.Duration    = time.Second
	ByteOrder       binary.ByteOrder = binary.LittleEndian
)

const (
	Limiter_bucket = 10
)

func init() {
}

func PrintLogo() {
	logx.DebugF("MaxConn: %d, MaxPacketSize: %d\n",
		GlobalObject.MaxConn,
		GlobalObject.MaxPacketSize)
}

func ReadWSMessage(conn *websocket.Conn, msg *Message) (int, error) {
	if conn == nil {
		return 0, errors.New(`ReadWSMessage conn is nil`)
	}
	if msg == nil {
		return 0, errors.New(`ReadWSMessage msg is nil`)
	}
	msg.Data = msg.Data[:0]
	msgType, reader, err := conn.NextReader()
	if err != nil {
		return 0, err
	}

	for {
		if len(msg.Data) >= cap(msg.Data)-1024 {
			newBuf := make([]byte, len(msg.Data), cap(msg.Data)*2)
			copy(newBuf, msg.Data)
			msg.Data = newBuf
		}

		n, err := reader.Read(msg.Data[len(msg.Data):cap(msg.Data)])
		if n > 0 {
			msg.Data = msg.Data[:len(msg.Data)+n]
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	msg.DataLen = uint32(len(msg.Data))

	return msgType, nil
}
