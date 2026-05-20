package gsframe

import (
	"github.com/tanenking/gsframe/internal/configx"
)

func ParseConfig(config_file string, config_value interface{}) (err error) {
	return configx.ParseConfig(config_file, config_value)
}
