package gsframe

import (
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/redisx"
)

func InitRedisHelper(configs *gsinf.RedisClusterConfig) error {
	return redisx.InitRedisHelper(configs)
}
func Redis() gsinf.IRedis {
	return redisx.Redis
}
func RedisError(err error) bool {
	return redisx.RedisError(err)
}
func RedisOK(ok string) bool {
	return redisx.RedisOK(ok)
}
