package redisx

import (
	"errors"
	"strings"

	"github.com/redis/go-redis/v9"
)

const (
	Status_OK = "OK"
)

func RedisError(err error) bool {
	if errors.Is(err, redis.Nil) {
		return false
	}
	return err != nil
}

func RedisOK(ok string) bool {
	return strings.Compare(strings.ToUpper(ok), Status_OK) == 0
}
