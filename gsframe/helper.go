package gsframe

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/mitchellh/mapstructure"
	uuid "github.com/satori/go.uuid"
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/helper"
	"github.com/tanenking/gsframe/internal/logx"
)

func IsNil(x interface{}) bool {
	if x == nil {
		return true
	}
	rv := reflect.ValueOf(x)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}

func ToJson(v interface{}) string {
	return helper.ToJson(v)
}

func FromJson(j string, out interface{}) error {
	return helper.FromJson(j, out)
}

func GetGlobalTimer() gsinf.ITimingWheel {
	return helper.GetGlobalTimer()
}

func GetCharactorCount(val string) (count int) {
	b := []rune(val)
	count = len(b)
	return
}
func GetChineseCharactorCount(val string) (count int) {
	for _, char := range val {
		if unicode.Is(unicode.Han, char) {
			count++
		} else {
			return 0
		}
	}
	return
}
func GenerateUUID(prefixs ...string) string {
	n := strings.Join(prefixs, "")
	ns := uuid.NewV4()
	u2 := uuid.NewV5(ns, n)
	return u2.String()
}
func Struct2Map(input interface{}) (out map[string]interface{}, err error) {
	out = map[string]interface{}{}
	err = mapstructure.Decode(input, &out)
	if err != nil {
		logx.ErrorF("Struct2Map err -> %v", err)
	}
	return
}

func AutoLock(lc *sync.Mutex) func() {
	if lc == nil {
		return func() {}
	}
	lc.Lock()
	return func() {
		lc.Unlock()
	}
}

// 数值转万分比小数
func GetTenThousandthRatio(num float64) float64 {
	return num * gsinf.TenThousandthRatio
}

func MakeUInt64(hi, lo uint32) uint64 {
	n := uint64(hi) << 32
	n += uint64(lo)
	return n
}
func ParseUInt64(n uint64) (hi uint32, lo uint32) {
	lo = uint32(n & 0xFFFFFFFF)
	hi = uint32(n >> 32)
	return
}
func MakeUInt32(hi, lo uint16) uint32 {
	n := uint32(hi) << 16
	n += uint32(lo)
	return n
}
func ParseUInt32(n uint32) (hi uint16, lo uint16) {
	lo = uint16(n & 0xFFFF)
	hi = uint16(n >> 16)
	return
}

// //////////////////////////////////////////////////////////////
func JsonArrayInterface2ArrayInt[T int | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64 | float32 | float64](i []interface{}) [][]T {
	if len(i) <= 0 {
		return [][]T{}
	}
	t := [][]T{}
	err := json.Unmarshal([]byte(ToJson(i)), &t)
	if err != nil {
		return t
	}
	return t
}

func Try(fun func()) {
	constants.Try(fun)
}
func TryHandle(fun func(), handler func(interface{})) {
	constants.TryHandle(fun, handler)
}
