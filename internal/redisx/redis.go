package redisx

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	uuid "github.com/satori/go.uuid"
	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/logx"
)

func (r *redis_t) AutoLock(key string, fn func()) {
	c := make(chan struct{})
	expt := time.Second * 30
	locker_value := uuid.NewV4()
	locker_value = uuid.NewV5(locker_value, key)
	constants.Go(func() {
		for {
			ok, _ := r.SetNX(context.Background(), key, locker_value.String(), expt).Result()
			if ok {
				c <- struct{}{}
				break
			}
			time.Sleep(time.Millisecond * 100) //100毫秒尝试一次
		}
	})
	//等待获取锁
	<-c

	defer func() {
		val, _ := r.Get(context.Background(), key).Result()
		if val == locker_value.String() {
			//查看,如果锁还是自己的,那就释放锁
			r.Del(context.Background(), key)
		}
	}()
	//执行逻辑
	fn()
}

func (r *redis_t) GetI(key string) (val int64, err error) {
	sval, err := r.Get(context.Background(), key).Result()
	if RedisError(err) {
		logx.ErrorF(`RedisGetI key %+v, err %+v`, key, err)
		return
	}
	if len(sval) <= 0 {
		return
	}
	val, err = strconv.ParseInt(sval, 10, 64)
	if err != nil {
		logx.ErrorF(`RedisGetI key %+v, err %+v`, key, err)
		return
	}
	return
}

func (r *redis_t) HGetI(key string, field string) (val int64, err error) {
	sval, err := r.HGet(context.Background(), key, field).Result()
	if RedisError(err) {
		logx.ErrorF(`RedisHGetI key %+v, field %+v, err %+v`, key, field, err)
		return
	}
	if len(sval) <= 0 {
		return
	}
	val, err = strconv.ParseInt(sval, 10, 64)
	if err != nil {
		logx.ErrorF(`RedisHGetI key %+v, err %+v`, key, err)
		return
	}
	return
}

/////////////////////////////////////////////////////////////////////////////////////

func (r *redis_t) CompileScript(script string) *redis.Script {
	if r.rdb_cluster != nil {
		logx.ErrorF(`redis lua script not allow run in cluster env`)
		return nil
	}
	return redis.NewScript(script)
}

func (r *redis_t) RunScript(script *redis.Script, keys []string, args ...interface{}) (ret []interface{}) {
	if r.rdb_cluster != nil {
		logx.ErrorF(`redis lua script not allow run in cluster env`)
		return nil
	}
	cmd := script.Run(context.Background(), r, keys, args...)
	rt, err := cmd.Result()
	if err != nil && err != redis.Nil {
		logx.ErrorF("RunScript, err = %v", err)
		return nil
	}
	ret, ok := rt.([]interface{})
	if !ok {
		return nil
	}
	return
}
