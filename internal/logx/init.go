package logx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"

	"github.com/sirupsen/logrus"
)

const (
	RunTime    gsinf.ContextKey = "runtime"
	_call_step int              = 3
)

type log_t struct {
	msg   string
	ctx   context.Context
	level logrus.Level
}

var (
	logInfo  *logrus.Logger
	fullpath string
	filename string
	inited   atomic.Int32
	buffer   chan *log_t
)

var logmsgPool = sync.Pool{
	New: func() interface{} {
		b := &log_t{}
		return b
	},
}

var (
	pid int
)

var NoLog bool = false
var JustPrint bool = false

const (
	ErrorLevel logrus.Level = iota + logrus.ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	MaxLevel
)
const (
	//文件保留2周
	file_WithMaxAge = time.Duration(time.Hour * 24 * 7 * 2)
	//文件一天一切换
	file_WithRotationTime = time.Duration(time.Hour * 24)
	//
	file_WithRotationCount = 0
	//
	file_WithRotationSize = -1
	//消息队列长度
	msg_queue_max_len = 40960
)

func init() {
	logInfo = logrus.New()
	fullpath = ""
	filename = "log.info"
	pid = os.Getpid()

	JustPrint = false

	initKernel32()

	buffer = make(chan *log_t, msg_queue_max_len)
}

func isLogRuntime() bool {
	log_runtime := os.Getenv(gsinf.Env_LogRuntime)
	log_runtime = strings.ToLower(log_runtime)
	if len(log_runtime) > 0 && log_runtime == "true" {
		return true
	}
	return false
}

func getLogPath() string {
	var _exists bool = false
	path := os.Getenv(gsinf.Env_LogPath)
	if len(path) <= 0 {
		return ""
		// p, _ := os.Getwd()
		// path = p + "/log/" + constants.ProjectName
	} else {
		if !filepath.IsAbs(path) {
			p, _ := os.Getwd()
			path = p + "/" + path
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

func getRuntime(step int) string {

	if !isLogRuntime() {
		return ""
	}

	_, file, line, ok := runtime.Caller(step) //_runtime_caller(step) //runtime.Caller(step)
	if !ok {
		return fmt.Sprintf("%s:%d", file, line)
	}
	fName := filepath.Base(file)
	return fmt.Sprintf("%s:%d", fName, line)
}

func GetLogContext(_call_step int) context.Context {
	ctx := context.WithValue(context.Background(), RunTime, getRuntime(_call_step))
	return ctx
}

func log(msg *log_t) {
	defer func() {
		logmsgPool.Put(msg)
	}()
	switch msg.level {
	case InfoLevel:
		colorPrint(msg.msg, green)
		logInfo.WithContext(msg.ctx).Infoln(msg.msg)
	case WarnLevel:
		colorPrint(msg.msg, yellow)
		logInfo.WithContext(msg.ctx).Warnln(msg.msg)
	case ErrorLevel:
		colorPrint(msg.msg, red)
		logInfo.WithContext(msg.ctx).Errorln(msg.msg)
	default:
		colorPrint(msg.msg, gray)
		logInfo.WithContext(msg.ctx).Debugln(msg.msg)
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
		DebugF(`log退出`)
	}()

	constants.AppExitWait()

	for {
		select {
		case <-constants.ExitChannel:
			return
		case msg, ok := <-buffer:
			if !ok || msg == nil {
				colorPrint(msg.msg, red)
				logrus.Errorln("log buffer 关闭")
			} else {
				log(msg)
			}
		}
	}
}
