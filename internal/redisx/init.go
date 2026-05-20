package redisx

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/logx"
)

type redis_t struct {
	redis.Cmdable
	rdb_cluster *redis.ClusterClient
	rdb_client  *redis.Client
}

var (
	Redis *redis_t
)

func init() {
	Redis = &redis_t{}
}

func initRedisCluster(addrs []string, user string, pwd string) error {
	Redis.rdb_cluster = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    addrs,
		Username: user,
		Password: pwd,
	})

	_, err := Redis.rdb_cluster.Ping(context.Background()).Result()
	return err
}

func initRedisSingleton(addr string, user string, pwd string) error {
	Redis.rdb_client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: user,
		Password: pwd,
		DB:       0,
	})

	_, err := Redis.rdb_client.Ping(context.Background()).Result()
	return err
}

func InitRedisHelper(configs *gsinf.RedisClusterConfig) error {
	if configs == nil || configs.Redis == nil || len(configs.Redis) <= 0 {
		return nil
	}

	Addrs := []string{}

	for _, v := range configs.Redis {
		if len(v.Host) <= 0 {
			logx.WarnF("redis config , host is nil")
			continue
		}
		if v.Port <= 0 {
			logx.WarnF("redis config , port is 0")
			continue
		}
		addr := fmt.Sprintf("%s:%d", v.Host, v.Port)
		Addrs = append(Addrs, addr)
	}

	if len(Addrs) <= 0 {
		return nil
	}

	hook := Prefix(configs.KeyPrefix)
	err := initRedisCluster(Addrs, configs.UserName, configs.Password)
	if err != nil {
		Redis.rdb_cluster = nil
		err = initRedisSingleton(Addrs[0], configs.UserName, configs.Password)
		if err != nil {
			logx.ErrorF("%v", err)
			return err
		}
		Redis.rdb_client.AddHook(hook)
		Redis.Cmdable = Redis.rdb_client
	} else {
		Redis.rdb_cluster.AddHook(hook)
		Redis.Cmdable = Redis.rdb_cluster
	}

	logx.InfoF("InitRedisHelper success")

	return nil
}
