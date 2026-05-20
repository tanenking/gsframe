package gsframe

import (
	"time"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/timex"
)

func InitTime(timeMilliOffset int64, timezone int32) {
	gsinf.TimeZoneLocation = time.FixedZone(`timezone`, int(timezone)*int(gsinf.Second_one_hour))
	timex.SetTimeMilliOffset(timeMilliOffset)
}
func GetNowTime() time.Time {
	return timex.GetNowTime()
}
func GetNowTimestamp() uint32 {
	return timex.GetNowTimestamp()
}
func GetNowTimestampMilli() int64 {
	return timex.GetNowTimestampMilli()
}

func SetTimeMilliOffset(offset int64) {
	timex.SetTimeMilliOffset(offset)
}

// 时间字符串转时间戳(默认local时间)
func LocalTimeStrToTimestamp(datetime string) uint32 {
	theTime, err := time.ParseInLocation(gsinf.TimeFormatString, datetime, gsinf.TimeZoneLocation)
	if err != nil {
		theTime, err = time.ParseInLocation(gsinf.TimeFormatStringShort, datetime, gsinf.TimeZoneLocation)
		if err != nil {
			return 0
		}
	}
	unixTime := uint32(theTime.Unix())
	return unixTime
}

// 获取指定时间的时间戳
func GetAppointDate(year int, month time.Month, day, hour, min, sec int) time.Time {
	t := time.Date(year, month, day, hour, min, sec, 0, gsinf.TimeZoneLocation)
	return t
}

// 获取当天0点时间戳
func GetToday0ClockTimestamp() uint32 {
	currentTime := GetNowTime()
	t := GetAppointDate(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0).Unix()
	return uint32(t)
}
