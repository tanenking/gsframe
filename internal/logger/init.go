package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	//消息队列长度
	msg_queue_max_len = 40960
)

type log_t struct {
	msg    string
	caller string
	level  zerolog.Level
}

var (
	zlogger       zerolog.Logger
	consoleWriter *zerolog.ConsoleWriter
	multiWriter   zerolog.LevelWriter
	log_caller    string
	fullpath      string
	filename      string = "log.info"
	inited        atomic.Int32
	buffer        chan *log_t = make(chan *log_t, msg_queue_max_len)
)

var logmsgPool = sync.Pool{
	New: func() interface{} {
		b := &log_t{}
		return b
	},
}

func Init() {
	if !inited.CompareAndSwap(0, 1) {
		return
	}

	log_caller = os.Getenv(gsinf.Env_LogRuntime)
	log_caller = strings.ToLower(log_caller)

	level := zerolog.DebugLevel
	if !constants.IsDebug() {
		log_level := os.Getenv(gsinf.Env_LogLevel)
		if len(log_level) > 0 {
			l, e := strconv.Atoi(log_level)
			if e == nil {
				level = zerolog.Level(l)
			}
		}
		if level < zerolog.DebugLevel || level >= zerolog.NoLevel {
			level = zerolog.InfoLevel
		}
	}

	var logFilePath = ""
	fullpath = getLogPath()
	if len(fullpath) > 0 {
		fullpath = filepath.Join(fullpath, constants.ServiceType)
		logFilePath = filepath.Join(fullpath, filename)
	}
	consoleWriter = &zerolog.ConsoleWriter{
		Out:          os.Stdout,
		TimeFormat:   time.DateTime,
		TimeLocation: gsinf.TimeZoneLocation,
		NoColor:      false,
		PartsOrder: []string{
			zerolog.CallerFieldName,
			zerolog.TimestampFieldName,
			zerolog.LevelFieldName,
			zerolog.MessageFieldName,
		},
	}
	if len(logFilePath) > 0 {
		logFile := lumberjack.Logger{
			Filename:   logFilePath, //日志文件路径
			MaxSize:    64,          //单个文件最大尺寸 (MB)
			MaxBackups: 512,         //最多保留的旧文件数量
			MaxAge:     28,          //旧文件最长保留天数
			// LocalTime:  true,
			Compress: false, //是否压缩旧文件为 .gz
		}
		textWriter := zerolog.ConsoleWriter{
			Out:          &logFile,
			TimeFormat:   time.DateTime,
			TimeLocation: gsinf.TimeZoneLocation,
			NoColor:      true,
			PartsOrder: []string{
				zerolog.CallerFieldName,
				zerolog.TimestampFieldName,
				zerolog.LevelFieldName,
				zerolog.MessageFieldName,
			},
		}
		multiWriter = zerolog.MultiLevelWriter(textWriter, consoleWriter)
	} else {
		multiWriter = zerolog.MultiLevelWriter(consoleWriter)
	}
	zlogger = zerolog.New(multiWriter).With().Timestamp().Logger().Level(level)

	constants.Go(func() { update() })
}

func GetLoggerWriter() io.Writer {
	return multiWriter
}

func getLogPath() string {
	var _exists bool = false
	path := os.Getenv(gsinf.Env_LogPath)
	if len(path) <= 0 {
		return ""
	} else {
		if !filepath.IsAbs(path) {
			p, _ := os.Getwd()
			path = filepath.Join(p, path)
		}
		var is bool
		is, _exists = constants.IsDir(path)
		if _exists && !is {
			os.Stderr.WriteString(`env log_path must need a dir`)
			// panic("env log_path must need a dir")
			return ""
		}
	}
	if !_exists {
		//创建文件夹
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Printf("%+v\n", err)
			// panic(err)
			return ""
		}
	}
	return path
}

func isLogRuntime() bool {
	if len(log_caller) > 0 && log_caller == "true" {
		return true
	}
	return false
}

func getCaller(step int) string {
	if !isLogRuntime() {
		return "unknow"
	}

	_, file, line, ok := runtime.Caller(step)
	if !ok {
		return fmt.Sprintf("%s:%d", file, line)
	}
	slist := strings.Split(file, "/")
	for idx, fname := range slist {
		if strings.HasPrefix(fname, "gsframe") {
			return fmt.Sprintf("%s:%d", filepath.Join(slist[idx:]...), line)
		}
	}

	return fmt.Sprintf("%s:%d", file, line)
}

func log(msg *log_t) {
	defer func() {
		logmsgPool.Put(msg)
	}()
	switch msg.level {
	case zerolog.DebugLevel:
		zlogger.Debug().Str("caller", msg.caller).Msg(msg.msg)
	case zerolog.InfoLevel:
		zlogger.Info().Str("caller", msg.caller).Msg(msg.msg)
	case zerolog.WarnLevel:
		zlogger.Warn().Str("caller", msg.caller).Msg(msg.msg)
	case zerolog.ErrorLevel:
		zlogger.Error().Str("caller", msg.caller).Msg(msg.msg)
	case zerolog.FatalLevel:
		zlogger.Fatal().Str("caller", msg.caller).Msg(msg.msg)
	case zerolog.PanicLevel:
		zlogger.Panic().Str("caller", msg.caller).Msg(msg.msg)
	default:
		zlogger.Log().Str("caller", msg.caller).Msg(msg.msg)
	}
}

func update() {
	defer func() {
		inited.Store(0)
		close(buffer)
		for msg := range buffer {
			log(msg)
		}
		constants.AppExitDone()
		Log().Debug(`log退出`)
	}()

	constants.AppExitWait()

	for {
		select {
		case <-constants.ExitChannel:
			return
		case msg, ok := <-buffer:
			if !ok || msg == nil {
				Log().Debug("log buffer 关闭")
			} else {
				log(msg)
			}
		}
	}
}
