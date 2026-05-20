package timex

import (
	"time"

	"github.com/tanenking/gsframe/gsinf"
)

var (
	timeMilliOffset int64 = 0
)

func init() {
	timeMilliOffset = 0
}

func GetNowTime() time.Time {
	return time.Now().Add(time.Millisecond * time.Duration(timeMilliOffset)).In(gsinf.TimeZoneLocation)
}
func GetNowTimestamp() uint32 {
	return uint32(GetNowTime().Unix())
}
func GetNowTimestampMilli() int64 {
	return GetNowTime().UnixMilli()
}

func SetTimeMilliOffset(offset int64) {
	timeMilliOffset = offset
}
