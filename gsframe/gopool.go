package gsframe

import (
	"github.com/tanenking/gsframe/internal/constants"
)

func Go(fn func()) {
	constants.Go(fn)
}

func Go2(fn func(params ...interface{}), params ...interface{}) {
	constants.Go2(fn, params...)
}
