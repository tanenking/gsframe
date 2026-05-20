package gsinf

import (
	"os"
	"strconv"
	"time"
)

const Second_one_day = uint32(24 * 3600)
const Millisecond_one_day = int64(Second_one_day) * 1000
const Second_one_hour = uint32(3600)
const Second_one_week = Second_one_day * 7

var TimeZoneLocation *time.Location

func parseTimezone() int {
	timezone_str := os.Getenv(Env_timezone)
	value, err := strconv.ParseInt(timezone_str, 10, 32)
	if err != nil {
		return 0
	}
	return int(value) * int(Second_one_hour)
}
func init() {
	TimeZoneLocation = time.FixedZone(`timezone`, parseTimezone())
}
