package gsframe

import (
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/cycledata"
)

func NewCycle(maxCount int32) gsinf.ICycle {
	return cycledata.NewCycle(maxCount)
}
