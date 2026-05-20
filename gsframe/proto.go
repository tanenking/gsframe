package gsframe

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/tanenking/gsframe/internal/helper"
	"github.com/tanenking/gsframe/internal/logx"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func GetProtoMessageName(protoFullName string) (name string) {
	items := strings.Split(protoFullName, ".")
	name = items[len(items)-1]
	return
}
func GetProtoMessagePrefixName(protoFullName string) (prefix string) {
	items := strings.Split(protoFullName, ".")
	prefix = strings.Join(items[0:len(items)-1], ".")
	return
}
func GetProtoMessageTypeByName(protoFullName string) protoreflect.MessageType {
	msgName := protoreflect.FullName(protoFullName)
	msgType, err := protoregistry.GlobalTypes.FindMessageByName(msgName)
	if err != nil {
		logx.ErrorF("GetProtoMessageTypeByName err: %v", err)
		return nil
	}
	return msgType
}
func NewProtoMessageByName(protoFullName string) (msg protoreflect.ProtoMessage, err error) {
	msgType := GetProtoMessageTypeByName(protoFullName)
	if msgType == nil {
		err = fmt.Errorf("can't find message type")
		logx.ErrorF("NewProtoMessageByName err: %v", err)
		return
	}
	msg = msgType.New().Interface()
	err = nil
	return
}
func MakeProtoMessage(protoFullName string, data []byte) protoreflect.ProtoMessage {
	msg, err := NewProtoMessageByName(protoFullName)
	if err != nil {
		return nil
	}
	if msg == nil {
		logx.ErrorF("msg = nil")
		return nil
	}
	err = proto.Unmarshal(data, msg)
	if err != nil {
		logx.ErrorF("err = %v", err)
		return nil
	}

	return msg
}
func MakeProtoMessage1(data []byte, out protoreflect.ProtoMessage) error {
	if reflect.ValueOf(out).Kind() != reflect.Pointer {
		return errors.New("MakeProtoMessage1 Error, out need pointer")
	}
	return proto.Unmarshal(data, out)
}
func MakeProtoMessage2[T protoreflect.ProtoMessage](protoFullName string, data []byte) (ret T, err error) {
	msg := MakeProtoMessage(protoFullName, data)
	if msg == nil {
		err = errors.New("MakeProtoMessage2 Error, protoFullName -> " + protoFullName)
		return
	}
	ret, ok := msg.(T)
	if !ok {
		return ret, errors.New("MakeProtoMessage2 Error, protoFullName -> " + protoFullName)
	}
	return ret, nil
}

// //////////////////////////////////////////////////////////////////////
func GetProtoMsgInfo(msg protoreflect.ProtoMessage) (name string, data []byte) {
	return helper.GetProtoMsgInfo(msg)
}
func GetProtoFullName(msg protoreflect.ProtoMessage) string {
	return helper.GetProtoFullName(msg)
}
func MarshalProto2Bytes(msg protoreflect.ProtoMessage, b []byte) (data []byte) {
	data, _ = proto.MarshalOptions{}.MarshalAppend(b, msg)
	return
}
