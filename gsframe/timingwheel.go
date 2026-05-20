package gsframe

import (
	"time"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/timingwheel"
)

func NewTimingWheel(tick time.Duration, wheelSize int64) gsinf.ITimingWheel {
	return timingwheel.NewTimingWheel(tick, wheelSize)
}
