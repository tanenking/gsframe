package logger_test

import (
	"testing"
	"time"

	"github.com/tanenking/gsframe/internal/logger"
)

func TestLog(t *testing.T) {
	logger.Init()
	logger.Log().Debug("init log test %+v", time.Duration(time.Hour))
	time.Sleep(time.Second * 5)
}
