

import (
	"testing"

	coconut "github.com/evelynocean/coconut/pb"
)

func TestUpdatePoints(t *testing.T) {
	conn, err := newRedisConnection("127.0.0.1:6379", 10, 100, 0)
	if err != nil {
		t.Error("newRedisConnection Error:", err)
	}

	s := &server{
		RedisClient: conn,
	}

	req := &coconut.PointsRequest{
		Level_1: "aaa",
		Level_2: "bbb",
		Level_3: "ccc",
	}

	for i := 0; i < 10; i++ {
		go func() {
			s.UpdatePoints(req)
		}()
	}

}
