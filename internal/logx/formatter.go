package logx

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/tanenking/gsframe/internal/timex"
)

const (
	red    = 4
	yellow = 6
	blue   = 1
	gray   = 8
	green  = 2
)

type formatter struct {
}

func (m *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	// timeformat := entry.Time.Format(gsinf.TimeFormatString)
	timeformat := timex.GetNowTime().Format(`2006-01-02 15:04:05.000`)

	_runtime := ""
	_pid := pid

	if entry.Context != nil {
		//_pid = entry.Context.Value("pid").(int)
		_runtime = entry.Context.Value(RunTime).(string)
		if len(_runtime) <= 0 {
			_runtime = "unknow"
		}
	}

	newLog := fmt.Sprintf("[%s] [%s][%d] [%s] [pid:%d] %s\n",
		entry.Level,
		timeformat,
		entry.Time.UnixMilli(),
		_runtime,
		_pid,
		entry.Message)

	b.WriteString(newLog)

	return b.Bytes(), nil
}
