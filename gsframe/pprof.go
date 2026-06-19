package gsframe

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // 这一行是关键！

	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/logger"
)

func RunPProf(port int32) {
	constants.Go(func() { pprof(port) })
}

func pprof(port int32) {
	if port <= 0 {
		return
	}
	a := fmt.Sprintf("localhost:%d", port)
	logger.Log().Debug("pprof start with :%s", a)
	err := http.ListenAndServe(a, nil)
	if err != nil {
		logger.Log().Debug("pprof err:%v", err)
		return
	}
}
