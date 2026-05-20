package helper

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func GetProtoMsgInfo(msg protoreflect.ProtoMessage) (name string, data []byte) {
	name = GetProtoFullName(msg)
	data, _ = proto.Marshal(msg)
	return
}
func GetProtoFullName(msg protoreflect.ProtoMessage) string {
	return string(proto.MessageName(msg))
}
