package gsframe

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
	"github.com/tanenking/gsframe/internal/logx"
)

func GetLogTraceFile() string {
	return logx.GetTraceFile()
}
func GetLogFullPath() string {
	return logx.GetLogFullPath()
}
func GetLogFileName() string {
	return logx.GetLogFileName()
}

func GetLoggerWriter() io.Writer {
	return logx.GetLoggerWriter()
}

func GetLogContext(_call_step int) context.Context {
	return logx.GetLogContext(_call_step)
}

func Log(ctx context.Context, level logrus.Level, msg string, args ...interface{}) {
	logx.Log(ctx, level, msg, args...)
}

func LogDebugF(msg string, args ...interface{}) {
	logx.Log(logx.GetLogContext(3), logrus.DebugLevel, msg, args...)
}
func LogInfoF(msg string, args ...interface{}) {
	logx.Log(logx.GetLogContext(3), logrus.InfoLevel, msg, args...)
}
func LogWarnF(msg string, args ...interface{}) {
	logx.Log(logx.GetLogContext(3), logrus.WarnLevel, msg, args...)
}
func LogErrorF(msg string, args ...interface{}) {
	logx.Log(logx.GetLogContext(3), logrus.ErrorLevel, msg, args...)
}
