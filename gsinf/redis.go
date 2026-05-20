package gsinf

import (
	"github.com/redis/go-redis/v9"
)

type IRedis interface {
	redis.Cmdable
	AutoLock(key string, fn func())
	GetI(key string) (val int64, err error)
	HGetI(key string, field string) (val int64, err error)
	CompileScript(script string) *redis.Script
	RunScript(script *redis.Script, keys []string, args ...interface{}) (ret []interface{})
	UpdateRankScore(key string, member interface{}, score float64)
	IncrbyRankScore(key string, member interface{}, score float64) (newScore float64)
	GetRankByMember(key string, member interface{}) redis.RankScore
}
