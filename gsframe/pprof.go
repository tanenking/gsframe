package gsframe

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // 这一行是关键！

	"github.com/tanenking/gsframe/internal/constants"
)

func RunPProf(port int32) {
	constants.Go(func() { pprof(port) })
}

func pprof(port int32) {
	if port <= 0 {
		return
	}
	a := fmt.Sprintf("localhost:%d", port)
	LogDebugF("pprof start with :%s", a)
	err := http.ListenAndServe(a, nil)
	if err != nil {
		LogDebugF("pprof err:%v", err)
		return
	}
}
