package coconut_redis

import (
	"encoding/json"
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

type LimitSetting struct {
	Level1 int
	Level2 int
	Level3 int
}

// set point
func PointSetBatch(conn *redis.Client, keys []string, point int, limitSetting map[string]int, expired int) (err error) {

	tmp := &LimitSetting{
		Level1: limitSetting["0"],
		Level2: limitSetting["1"],
		Level3: limitSetting["2"],
	}
	/**
		    local point = ARGV[1]
			local limit = cjson.decode(ARGV[2])
			local expired = ARGV[3]

			if(redis.call('GET',KEYS[1]) + point <= tonumber(limit.Level1)) then
				return redis.call('INCRBY',KEYS[1], point)
			else
				return tostring(-99)
			end

	if(redis.call('GET',KEYS[2]) + point <= tonumber(limit.Level2)) then
		return redis.call('INCRBY',KEYS[2], point)
	else
		return tostring(-98)
	end

	if(redis.call('GET',KEYS[3]) + point <= tonumber(limit.Level3)) then
		return redis.call('INCRBY',KEYS[3], point)
	else
		return tostring(-97)
	end
	**/
	luaScript := `
	local point = tonumber(ARGV[1])
	local limit = cjson.decode(ARGV[2])
	local expired = tonumber(ARGV[3])

	-- 先GET一次KEY, 沒有KEY的要SET
	if( redis.call('GET', KEYS[1]) == nil or redis.call('GET', KEYS[1]) == false) then
		redis.call('SETEX', KEYS[1], expired, 0)
	end

	if( redis.call('GET', KEYS[2]) == nil or redis.call('GET', KEYS[2]) == false) then
		redis.call('SETEX', KEYS[2], expired, 0)
	end

	if( redis.call('GET', KEYS[3]) == nil or redis.call('GET', KEYS[3]) == false) then
		redis.call('SETEX', KEYS[3], expired, 0)
	end

	if(redis.call('GET', KEYS[1]) + point <= tonumber(limit.Level1)) then
		redis.call('INCRBY',KEYS[1], point)
	else
		return tostring(-99)
	end

	if(redis.call('GET', KEYS[2]) + point <= tonumber(limit.Level2)) then
		redis.call('INCRBY',KEYS[2], point)
	else
		return tostring(-98)
	end

	if(redis.call('GET', KEYS[3]) + point <= tonumber(limit.Level3)) then
		redis.call('INCRBY',KEYS[3], point)
	else
		return tostring(-97)
	end

	return 'ok'
	`
	// fmt.Println(" ------- keys:", keys)
	// fmt.Println(" ------- point:", point)
	// fmt.Println(" ------- tmp:", tmp.Level1, ", 2:", tmp.Level2, ", 3:", tmp.Level3)
	// fmt.Println(" ------- expired:", expired)
	script, err := conn.ScriptLoad(luaScript).Result()
	if err != nil {
		return err
	}

	_, err = conn.EvalSha(script, keys, point, tmp.MarshalBinary(), expired).Result()
	// fmt.Println("reply:", reply, ", err:", err)
	if err != nil {
		return err
	}

	return
}

func (s *LimitSetting) MarshalBinary() (ret string) {
	data, _ := json.Marshal(s)
	ret = string(data)
	return ret
}

func (s *LimitSetting) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
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
