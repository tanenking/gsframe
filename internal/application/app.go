package application

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/helper"
	"github.com/tanenking/gsframe/internal/logx"
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

	logx.InitLogx()

	logo = ``
	_logo := logo + "\n"
	_logo += topLine + "\n"
	_logo += fmt.Sprintf("%s [coder] martin                                    %s", borderLine, borderLine) + "\n"
	_logo += fmt.Sprintf("%s [time] 2022-01-29                                 %s", borderLine, borderLine) + "\n"
	_logo += bottomLine + "\n"

	logx.InfoF("%s", _logo)

	startTime = time.Now()

	return true
}

func ProgramRunning() {
	defer func() {
		logx.InfoF("主进程退出 server close pid = %d", os.Getpid())
		notifySubKill()
		close(constants.ExitChannel)
		constants.ExitWaitGroup.Wait()
	}()

	logx.InfoF("服务启动成功")
	if !writePid() {
		return
	}

	helper.GetGlobalTimer().Start()
	for sig := range constants.GetSignals() {
		logx.InfoF("signal receive: %v", sig)
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
		logx.ErrorF(`StartProcess err => %+v`, err)
		return
	}
	sub_pids = append(sub_pids, ps)
}
