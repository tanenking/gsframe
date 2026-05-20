package gsframe

import (
	"os"

	"github.com/tanenking/gsframe/internal/application"
	"github.com/tanenking/gsframe/internal/constants"
)

func GetRunningSeconds() uint32 {
	return application.GetRunningSeconds()
}

func InitProgram(project_name string, service_type string) bool {
	return application.InitProgram(project_name, service_type)
}

func ProgramRunning() {
	application.ProgramRunning()
}

func AppExitWait() chan struct{} {
	constants.AppExitWait()
	return constants.ExitChannel
}
func AppExitDone() {
	constants.AppExitDone()
}

func StartSubProcess(cmd string, args []string, attr *os.ProcAttr) {
	application.StartSubProcess(cmd, args, attr)
}
