package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	coconutError "github.com/evelynocean/coconut/lib/error"
	coconut_model "github.com/evelynocean/coconut/model"
	coconut "github.com/evelynocean/coconut/pb"
	coconut_redis "github.com/evelynocean/coconut/redis"
)

func (s *server) Ping(ctx context.Context, in *coconut.PingRequest) (r *coconut.Pong, err error) {
	r = &coconut.Pong{
		Pong: "pong",
	}

	return r, err
}

func (s *server) UpdatePoints(ctx context.Context, in *coconut.PointsRequest) (r *coconut.RetResultStatus, err error) {
	start := time.Now()

	defer func() {
		Logger.WithFields(map[string]interface{}{
			"input":        in,
			"execute_time": time.Since(start).Seconds(),
			"response":     r,
		}).Debugf("test")

		HandlerPanicRecover(&err)
	}()

	// TODO: input data check
	if in.Level_1 == "" || in.Level_2 == "" || in.Level_3 == "" {
		return nil, coconutError.ParseError(coconutError.ErrServer, errors.New("invalid parameter"))
	}

	r = &coconut.RetResultStatus{}
	sets := &coconut_redis.KeySet{
		Level1:   in.Level_1,
		Level2:   in.Level_2,
		Level3:   in.Level_3,
		UserName: in.UserName,
	}
	keys := coconut_redis.GetPointKey(sets)

	limitSettings, err := coconut_model.LimitSQL.GetLimit()
	// fmt.Println("::::::: ", config.Environment, " :::::::: ", limitSettings)
	reply, err := coconut_redis.PointSetBatch(s.RedisClient, keys, int(in.Point), limitSettings, 30)
	if err != nil {
		Logger.WithFields(map[string]interface{}{
			"test": 111,
			"time": time.Now().UnixNano(),
			"err:": err.Error(),
		}).Errorf("testError")
		return nil, coconutError.ParseError(coconutError.ErrRedis, err)
	}

	if reply != "ok" {
		fmt.Println(" 踩到限額 發TG", reply, ", err:", err)

		// msg := tgBot.NewMessage(TgBotChatID, fmt.Sprintf("踩到限額: %s", reply))
		// _, err := s.Bot.Send(msg)
		// if err != nil {
		// 	Logger.WithFields(map[string]interface{}{
		// 		"err:": err.Error(),
		// 	}).Errorf("Bot.Send")
		// }

	}
	// NSQ發送 各層級
	var data []*coconut.PointInfo
	reg := regexp.MustCompile(`^LEVEL.*:(\w+)$`)

	for _, v := range keys {
		resultPoint, err := coconut_redis.PointGet(s.RedisClient, v)
		if err != nil {
			return nil, coconutError.ParseError(coconutError.ErrRedis, err)
		}
		if resultPoint > 0 {
			matchSlice := reg.FindStringSubmatch(v)
			d := &coconut.PointInfo{
				Name:   matchSlice[1],
				Points: int32(resultPoint),
			}
			data = append(data, d)
		}
	}

	str, err := json.Marshal(data)
	fmt.Println(":::::: str:", string(str), ", err:", err)
	if err != nil {
		return nil, coconutError.ParseError(coconutError.ErrServer, err)
	}

	err = s.NsqProducer.Publish("COCONUT_UPDATE_POINT", []byte(str))
	if err != nil {
		return nil, coconutError.ParseError(coconutError.ErrServer, err)
	}

	r = &coconut.RetResultStatus{
		Success: true,
	}

	return r, nil
}

func (s *server) GetPoints(ctx context.Context, in *coconut.GetPointsRequest) (r *coconut.RetPoints, err error) {
	start := time.Now()

	var data []*coconut.PointInfo

	defer func() {
		Logger.WithFields(map[string]interface{}{
			"input":        in,
			"execute_time": time.Since(start).Seconds(),
			"response":     r,
		}).Debugf("test")

		HandlerPanicRecover(&err)
	}()

	if in.Level_1 == "" || in.Level_2 == "" || in.Level_3 == "" {
		return nil, coconutError.ParseError(coconutError.ErrServer, errors.New("invalid parameter"))
	}

	reg := regexp.MustCompile(`^LEVEL.*:(\w+)$`)

	sets := &coconut_redis.KeySet{
		Level1: in.Level_1,
		Level2: in.Level_2,
		Level3: in.Level_3,
	}
	keys := coconut_redis.GetPointKey(sets)

	for _, v := range keys {
		resultPoint, err := coconut_redis.PointGet(s.RedisClient, v)
		if err != nil {
			return nil, coconutError.ParseError(coconutError.ErrRedis, err)
		}
		if resultPoint > 0 {
			matchSlice := reg.FindStringSubmatch(v)
			d := &coconut.PointInfo{
				Name:   matchSlice[1],
				Points: int32(resultPoint),
			}
			data = append(data, d)
		}
	}

	r = &coconut.RetPoints{
		Data: data,
	}

	return r, nil
}
