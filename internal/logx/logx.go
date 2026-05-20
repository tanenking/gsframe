package logx

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

// level ErrorLevel-DebugLevel
func InitLogx() {

	if !inited.CompareAndSwap(0, 1) {
		return
	}

	level := DebugLevel
	if !constants.IsDebug() {
		log_level := os.Getenv(gsinf.Env_LogLevel)
		if len(log_level) > 0 {
			l, e := strconv.Atoi(log_level)
			if e == nil {
				level = logrus.Level(l)
			}
		}
		if level < ErrorLevel || level >= MaxLevel {
			level = InfoLevel
		}
	}

	var mw_info io.Writer

	fullpath = getLogPath()
	if len(fullpath) > 0 {
		fullpath += "/" + constants.ServiceType

		filename = "log.info"
		file_info := fullpath + "/" + filename

		writer_info, _ := rotatelogs.New(
			file_info+".%Y%m%d%H%M",
			rotatelogs.WithLinkName(file_info),
			rotatelogs.WithMaxAge(file_WithMaxAge),
			rotatelogs.WithRotationCount(file_WithRotationCount),
			rotatelogs.WithRotationTime(file_WithRotationTime),
			rotatelogs.WithRotationSize(file_WithRotationSize),
		)
		mw_info = io.MultiWriter(writer_info, os.Stdout)
	} else {
		mw_info = io.MultiWriter(os.Stdout)
	}
	logInfo.SetOutput(mw_info)

	logInfo.SetLevel(level)

	formatter := &formatter{
		//ForceColors:               true,
		//EnvironmentOverrideColors: true,
		//FullTimestamp:   true,
		//TimestampFormat: constants.TimeFormatString,
		//DisableSorting:            true,
		//DisableLevelTruncation:    true,
		//PadLevelText: true,
	}

	logInfo.SetFormatter(formatter)

	constants.Go(func() { update() })
}

func GetTraceFile() string {
	return fullpath + "/trace"
}
func GetLogFullPath() string {
	return fullpath
}
func GetLogFileName() string {
	return filename
}

func GetLoggerWriter() io.Writer {
	return logInfo.Out
}

func Log(ctx context.Context, level logrus.Level, msg string, args ...interface{}) {
	defer constants.AutoRecover()()
	if NoLog {
		return
	}
	if level > logInfo.Level {
		return
	}

	t := logmsgPool.Get().(*log_t)
	t.msg = fmt.Sprintf(msg, args...)
	t.ctx = ctx
	t.level = level

	if inited.Load() <= 0 || JustPrint {
		// os.Stdout.WriteString(fmt.Sprintf("["+level.String()+"]"+msg+"\n", args...))
		log(t)
		return
	}

	buffer <- t
}

// func TraceBack() {
// 	if NoLog {
// 		return
// 	}
// 	vmsg := string(debug.Stack())
// 	ctx := GetLogContext(_call_step)
// 	Log(ctx, DebugLevel, vmsg)
// }

func DebugF(msg string, args ...interface{}) {
	if NoLog {
		return
	}
	ctx := GetLogContext(_call_step)
	// Log(ctx, DebugLevel, fmt.Sprintf(msg, args...))
	Log(ctx, DebugLevel, msg, args...)
}

func InfoF(msg string, args ...interface{}) {
	if NoLog {
		return
	}
	ctx := GetLogContext(_call_step)
	// Log(ctx, InfoLevel, fmt.Sprintf(msg, args...))
	Log(ctx, InfoLevel, msg, args...)
}

func WarnF(msg string, args ...interface{}) {
	if NoLog {
		return
	}
	ctx := GetLogContext(_call_step)
	// Log(ctx, WarnLevel, fmt.Sprintf(msg, args...))
	Log(ctx, WarnLevel, msg, args...)
}

func ErrorF(msg string, args ...interface{}) {
	if NoLog {
		return
	}
	ctx := GetLogContext(_call_step)
	// Log(ctx, ErrorLevel, fmt.Sprintf(msg, args...))
	Log(ctx, ErrorLevel, msg, args...)
}
