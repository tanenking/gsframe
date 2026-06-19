package logger

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
)

func Logger() *log_t {
	t := logmsgPool.Get().(*log_t)
	return t
}

func Log() *log_t {
	t := logmsgPool.Get().(*log_t)._caller(3)
	return t
}

func (r *log_t) _caller(caller_skip int) *log_t {
	r.caller = getCaller(caller_skip)
	return r
}

func (r *log_t) Caller(caller_skip int) gsinf.ILogger {
	r.caller = getCaller(caller_skip)
	return r
}

func (r *log_t) Debug(msg string, args ...interface{}) {
	r.log(zerolog.DebugLevel, msg, args...)
}

func (r *log_t) Info(msg string, args ...interface{}) {
	r.log(zerolog.DebugLevel, msg, args...)
}

func (r *log_t) Warn(msg string, args ...interface{}) {
	r.log(zerolog.DebugLevel, msg, args...)
}

func (r *log_t) Error(msg string, args ...interface{}) {
	r.log(zerolog.DebugLevel, msg, args...)
}

func (r *log_t) Fatal(msg string, args ...interface{}) {
	r.log(zerolog.DebugLevel, msg, args...)
}

func (r *log_t) Panic(msg string, args ...interface{}) {
	r.log(zerolog.DebugLevel, msg, args...)
}

func (r *log_t) log(level zerolog.Level, msg string, args ...interface{}) {
	defer constants.AutoRecover()()
	if level < zlogger.GetLevel() {
		logmsgPool.Put(r)
		return
	}
	r.msg = fmt.Sprintf(msg, args...)
	r.level = level

	if inited.Load() <= 0 {
		log(r)
		return
	}

	buffer <- r
}
