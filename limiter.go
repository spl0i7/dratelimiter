package dratelimiter

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"strconv"
	"time"
)

const RateLimitKey = "__RATELIMIT__"
const RequestSet = "__REQUESTSET__"



type Limiter struct {
	RedisClient *redis.Client
	RateLimit   int64
}

func (l *Limiter) NewLimiter(r *redis.Client, rate int64) error {
	l.RateLimit = rate
	l.RedisClient = r

	err := r.Ping().Err()
	if err != nil { return err }

	v, err := r.Get(RateLimitKey).Result()
	if err == redis.Nil {
		// key not set
		r.Set(RateLimitKey, rate, 0)
	} else {
		// global rate limit is already set
		rateLimit, err := strconv.Atoi(v)

		if err != nil { return err }

		if int64(rateLimit) != rate {
			log.Printf("global rate limit already set to %d \n", rateLimit)
			return errors.New("GlobalLimitSet")
		}
	}

	return nil
}

func (l *Limiter) Allow() bool {

	txf := func(tx *redis.Tx) error {

		timestamp := time.Now().UnixNano()
		_, err := tx.Pipelined(func(pipe redis.Pipeliner) error {
			_, err := tx.ZRemRangeByScore(RequestSet, "0", fmt.Sprintf("%g", float64(timestamp - 1e9))).Result()
			if err != nil {
				return err
			}
			res, err := tx.ZCard(RequestSet).Result()
			if err != nil {
				return err
			}

			if res >= l.RateLimit {
				return errors.New("limit exceed")
			}

			tx.ZAdd(RequestSet, redis.Z{ Score : float64(timestamp), Member: fmt.Sprintf("%d", timestamp)})

			return nil
		})
		return err
	}

	err := l.RedisClient.Watch(txf, RequestSet)
	//if err != nil {
	//	log.Println(err)
	//}

	return err == nil

}

func (l *Limiter) ResetGlobalLimit(rate int64) error{
	_, err := l.RedisClient.Set(RateLimitKey, rate, 0).Result()
	return err
}

func (l *Limiter) Done()  {
	_, _  = l.RedisClient.Del(RateLimitKey).Result()
	_, _  = l.RedisClient.Del(RequestSet).Result()
	_ = l.RedisClient.Close()
	l.RedisClient = nil

}