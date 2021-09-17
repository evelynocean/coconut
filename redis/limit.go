package coconut_redis

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type KeySet struct {
	Level1   string
	Level2   string
	Level3   string
	UserName string
}

// get redis keys
func GetPointKey(keyset *KeySet) (keys []string) {
	keys = append(keys, fmt.Sprintf("%s:%s", "LEVEL", keyset.Level1))
	keys = append(keys, fmt.Sprintf("%s:%s:%s", "LEVEL", keyset.Level1, keyset.Level2))
	keys = append(keys, fmt.Sprintf("%s:%s:%s:%s", "LEVEL", keyset.Level1, keyset.Level2, keyset.Level3))
	if keyset.UserName != "" {
		keys = append(keys, fmt.Sprintf("%s:%s:%s:%s:%s", "LEVEL", keyset.Level1, keyset.Level2, keyset.Level3, keyset.UserName))
	}

	return keys
}

// set point
func PointSet(conn *redis.Client, key string, value int, expired time.Duration, limit int) (err error) {
	reply, err := conn.Get(key).Int()
	if !errors.Is(err, redis.Nil) {
		return
	}

	if reply+value > limit {
		return errors.New("Over limit")
	}
	_, err = conn.IncrBy(key, int64(value)).Result()
	if err != nil {
		return
	}

	_, err = conn.Expire(key, expired).Result()
	if err != nil {
		// log, but return nil
		return nil
	}

	return
}

// get point
func PointGet(conn *redis.Client, key string) (reply int, err error) {
	reply, err = conn.Get(key).Int()
	if err != nil {
		return
	}

	return
}
