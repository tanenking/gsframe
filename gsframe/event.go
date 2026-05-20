package gsframe

import (
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
)

func NewEventManager() gsinf.IEventManager {
	return constants.NewEventManager()
}
