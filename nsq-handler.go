package main

import (
	"log"
	"sync"
	"time"

	nsq "github.com/nsqio/go-nsq"
)

type NsqConsumerList struct {
	sync.Map
}

// Set 存Consumer
func (ncl *NsqConsumerList) Set(key int64, c *nsq.Consumer) *NsqConsumerList {
	ncl.Store(key, c)
	return ncl
}

// Each callback每個Consumer
func (ncl *NsqConsumerList) Each(callback func(c *nsq.Consumer) error) error {
	ncl.Range(func(k, v interface{}) bool {
		c, ok := v.(*nsq.Consumer)

		if !ok {
			return false
		}

		err := callback(c)

		if err != nil {
			return false
		}

		return true
	})

	return nil
}

func TestNSQConsumer() nsq.Handler {
	return nsq.HandlerFunc(func(message *nsq.Message) (err error) {
		message.DisableAutoResponse()
		defer message.Finish()
		log.Printf("========================== 收到的訊息是：%v", string(message.Body))
		time.Sleep(time.Duration(20) * time.Second)
		log.Printf("========= 故意睡 20s 測試是否有等consumer處理完才結束")
		return nil
	})
}
