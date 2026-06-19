package gsframe

import (
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/logger"
)

// func LogDebugF(msg string, args ...interface{}) {
// 	logger.Log().Debug(msg, args...)
// }
// func LogInfoF(msg string, args ...interface{}) {
// 	logger.Log().Info(msg, args...)
// }
// func LogWarnF(msg string, args ...interface{}) {
// 	logger.Log().Warn(msg, args...)
// }
// func LogErrorF(msg string, args ...interface{}) {
// 	logger.Log().Error(msg, args...)
// }

func Logger() gsinf.ILogger {
	return logger.Logger()
}
