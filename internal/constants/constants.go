package constants

import (
	"context"
	"net"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/tanenking/gsframe/gsinf"
	"google.golang.org/grpc/peer"
)

var (
	//服务状态
	system_status gsinf.SystemStatus_t
)

func SetSystemStatus(status gsinf.SystemStatus_t) {
	system_status = status
}

func GetSystemStatus() gsinf.SystemStatus_t {
	return system_status
}

// /////////////////////////////////////////////////////////////////////////////
func GetSystem() string {
	return runtime.GOOS
}

func GetServiceMode() string {
	mode := strings.ToUpper(os.Getenv(gsinf.Env_ServiceMode))
	if len(mode) <= 0 || !IsValidServiceMode(mode) {
		mode = gsinf.ServiceMode_TEST
	}
	return mode
}

func IsValidServiceMode(mode string) bool {
	_, ok := Service_mode_map[mode]
	return ok
}

func IsCoredump() bool {
	coredump_mode := os.Getenv(gsinf.Env_Coredump)
	return len(coredump_mode) > 0
}

func IsDebug() bool {
	runtime_mode := os.Getenv(gsinf.Env_RuntimeMode)
	return strings.ToLower(runtime_mode) == gsinf.RuntimeMode_Debug
}

func GetServiceHost() string {
	return ServiceHost
}

func GetPeerAddr(ctx context.Context) string {
	var addr string
	if pr, ok := peer.FromContext(ctx); ok {
		if tcpAddr, ok := pr.Addr.(*net.TCPAddr); ok {
			addr = tcpAddr.IP.String()
		} else {
			addr = pr.Addr.String()
		}
	}
	return addr
}

func ProgramExit() {
	signals <- syscall.SIGTERM
}

func GetSignals() chan os.Signal {
	return signals
}
