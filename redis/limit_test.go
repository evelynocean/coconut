package coconut_redis

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

// newRedisConnection ...
func newRedisConnection() (client *redis.Client, err error) {
	client = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   2,
	})

	_, err = client.Ping().Result()

	return
}

func TestGetPointKey(t *testing.T) {
	sets := &KeySet{
		Level1:   "AA",
		Level2:   "B",
		Level3:   "CCC",
		UserName: "Coconut",
	}

	keys := GetPointKey(sets)
	if keys[0] != fmt.Sprintf("%s:%s", "LEVEL", sets.Level1) {
		t.Error("TestGetPointKey 0 Error")
	}

	if keys[1] != fmt.Sprintf("%s:%s:%s", "LEVEL", sets.Level1, sets.Level2) {
		t.Error("TestGetPointKey 1 Error")
	}

	if keys[2] != fmt.Sprintf("%s:%s:%s:%s", "LEVEL", sets.Level1, sets.Level2, sets.Level3) {
		t.Error("TestGetPointKey 2 Error")
	}

	if keys[3] != fmt.Sprintf("%s:%s:%s:%s:%s", "LEVEL", sets.Level1, sets.Level2, sets.Level3, sets.UserName) {
		t.Error("TestGetPointKey 3 Error")
	}
}

func TestPoint(t *testing.T) {
	conn, err := newRedisConnection()
	if err != nil {
		t.Error("newRedisConnection Error:", err)
	}

	sets := &KeySet{
		Level1:   "AA",
		Level2:   "B",
		Level3:   "CCC",
		UserName: "Coconut",
	}

	keys := GetPointKey(sets)

	for _, v := range keys {
		originPoint, err := PointGet(conn, v)
		if err != nil {
			if err != redis.Nil {
				t.Error("PointGet origin Error:", err)
			}
		}

		err = PointSet(conn, v, 100, time.Duration(30)*time.Second, 500)
		if err != nil {
			t.Error("PointSet Error:", err)
		}

		resultPoint, err := PointGet(conn, v)
		if err != nil {
			t.Error("PointGet result Error:", err)
		}

		if resultPoint != originPoint+100 {
			t.Error("Set point Error")
		}
	}

}
