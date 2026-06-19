package helper

import (
	"math/rand"
	"time"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/timingwheel"
)

const ()

var (
	globalTimer gsinf.ITimingWheel //全局定时器

	// filterManager *filter.DirtyManager
)

func init() {
	globalTimer = timingwheel.NewTimingWheel(time.Millisecond*100, 600)
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

// func InitFilterManager(filterFile string) {
// 	memStore, err := store.NewMemoryStore(store.MemoryConfig{
// 		DataSource: []string{},
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	filterManager = filter.NewDirtyManager(memStore)
// }

// // 铭感词过滤
// func FilterVerify(text string) bool {
// 	if filterManager == nil {
// 		return false
// 	}
// 	result, err := filterManager.Filter().Filter(text)
// 	if err != nil {
// 		logger.Log().Error("FilterVerify err -> %v", err)
// 		return false
// 	}
// 	return result == nil
// }

func GetGlobalTimer() gsinf.ITimingWheel {
	return globalTimer
}
