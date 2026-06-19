package application

import (
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/helper"
	"github.com/tanenking/gsframe/internal/logger"
)

var startTime time.Time

func GetRunningSeconds() uint32 {
	return uint32(time.Since(startTime).Seconds())
}

func InitProgram(project_name string, service_type string) bool {
	runtime.GOMAXPROCS(runtime.NumCPU())
	constants.ProjectName = project_name
	constants.ServiceType = service_type
	setTitle(service_type)

	logger.Init()

	startTime = time.Now()

	return true
}

func ProgramRunning() {
	defer func() {
		logger.Log().Info("主进程退出 server close pid = %d", os.Getpid())
		notifySubKill()
		close(constants.ExitChannel)
		constants.ExitWaitGroup.Wait()
	}()

	logger.Log().Info("服务启动成功")
	if !writePid() {
		return
	}

	helper.GetGlobalTimer().Start()
	for sig := range constants.GetSignals() {
		logger.Log().Info("signal receive: %v", sig)
		switch sig {
		case syscall.SIGINT:
			return
		case syscall.SIGTERM:
			return
		}
	}
}

func StartSubProcess(cmd string, args []string, attr *os.ProcAttr) {
	ps, err := os.StartProcess(cmd, args, attr)
	if err != nil {
		logger.Log().Error(`StartProcess err => %+v`, err)
		return
	}
	sub_pids = append(sub_pids, ps)
}
