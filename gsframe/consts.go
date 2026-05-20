package gsframe

import (
	"context"
	"os"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
)

func SetSystemStatus(status gsinf.SystemStatus_t) {
	constants.SetSystemStatus(status)
}

func GetSystemStatus() gsinf.SystemStatus_t {
	return constants.GetSystemStatus()
}

// /////////////////////////////////////////////////////////////////////////////
func GetSystem() string {
	return constants.GetSystem()
}

func GetServiceMode() string {
	return constants.GetServiceMode()
}

func IsValidServiceMode(mode string) bool {
	return constants.IsValidServiceMode(mode)
}

func IsCoredump() bool {
	return constants.IsCoredump()
}

func IsDebug() bool {
	return constants.IsDebug()
}

func GetServiceHost() string {
	return constants.GetServiceHost()
}

func GetPeerAddr(ctx context.Context) string {
	return constants.GetPeerAddr(ctx)
}

func ProgramExit() {
	constants.ProgramExit()
}

func GetSignals() chan os.Signal {
	return constants.GetSignals()
}

func GetProjectName() string {
	return constants.ProjectName
}

func GetServiceType() string {
	return constants.ServiceType
}

func IsWindowsSystem() bool {
	return constants.IsWindowsSystem()
}
