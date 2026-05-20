package redisx

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func (r *redis_t) UpdateRankScore(key string, member interface{}, score float64) {
	r.ZAdd(context.Background(), key, redis.Z{Score: score, Member: member})
}

func (r *redis_t) IncrbyRankScore(key string, member interface{}, score float64) (newScore float64) {
	return r.ZIncrBy(context.Background(), key, score, fmt.Sprintf(`%v`, member)).Val()
}

func (r *redis_t) GetRankByMember(key string, member interface{}) redis.RankScore {
	// ret, err := r.ZRevRankWithScore(context.Background(), key, fmt.Sprintf(`%v`, member)).Result()
	// if err != nil {
	// 	return redis.RankScore{Rank: -1, Score: 0}
	// }
	ret := redis.RankScore{Rank: -1, Score: 0}
	nRank, err := r.ZRevRank(context.Background(), key, fmt.Sprintf(`%v`, member)).Result()
	if err != nil {
		return ret
	}
	ret.Rank = nRank

	ret.Score = r.ZScore(context.Background(), key, fmt.Sprintf(`%v`, member)).Val()

	return ret
}
