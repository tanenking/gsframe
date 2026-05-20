package gsinf

import "reflect"

type (
	ContextKey string
)

const (
	RpcSendRecvMaxSize = 1024 * 1024 * 4
)

const (
	ServiceMode_TEST = "TEST"
	ServiceMode_PROD = "PROD"
)

const (
	RuntimeMode_Debug   = "debug"
	RuntimeMode_Release = "release"
)

const (
	Env_ServiceHost = "service_host"
	Env_Gopath      = "GOPATH"
	Env_RuntimeMode = "runtime_mode"
	Env_ServiceMode = "service_mode"
	Env_Coredump    = "coredump"
	Env_LogLevel    = "log_level"
	Env_LogPath     = "log_path"
	Env_LogRuntime  = "log_runtime"
	Env_PidPath     = "pid_path"
	Env_timezone    = "timezone"
)

const (
	TimeFormatString      = "2006-01-02 15:04:05"
	TimeFormatStringShort = "2006-01-02"
	DateNumberYMD         = "20060102"
	DateNumberYM          = "200601"
)

const (
	TenThousandthRatio = 0.0001
)

type SystemStatus_t int32

const (
	SystemStatus_Normal   SystemStatus_t = iota
	SystemStatus_Maintain                //维护
)

func IsNil(x interface{}) bool {
	if x == nil {
		return true
	}
	rv := reflect.ValueOf(x)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}
