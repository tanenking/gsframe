package gsframe

import "github.com/tanenking/gsframe/internal/constants"

func AutoRecover() func() {
	return constants.AutoRecover()
}
