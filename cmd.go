package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	nsq "github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
	"github.com/syhlion/gocql"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	coconut_model "github.com/evelynocean/coconut/model"
	coconut "github.com/evelynocean/coconut/pb"
)

type server struct {
	ScyllaSession *gocql.Session
	RedisClient   *redis.Client
	NsqProducer   *nsq.Producer
}

func start(c *cli.Context) {
	envInit(c)

	var (
		err              error
		shutdownObserver = make(chan os.Signal, 1)
	)

	defer func() {
		if err != nil {
			Logger.WithFields(map[string]interface{}{
				"err:": err.Error(),
			}).Errorf("service start error")
		}
	}()

	redisClient, err := newRedisConnection(config.RedisAddr, config.RedisMaxIdle, config.RedisMaxConn, config.RedisNo)
	if err != nil {
		return
	}

	// set scylla connection
	session, err := newCQLSession()
	if err != nil {
		return
	}
	// 要有取到連線 session.Close才不會噴錯
	defer session.Close()

	// nsq producer
	nsqProducer, err := newNSQProducer(config.NsqdAddr)
	if err != nil {
		return
	}

	defer func() {
		nsqProducer.Stop()
	}()

	if config.Environment == "local" {
		coconut_model.InitMock()
	} else {
		coconut_model.Init(session)
	}

	sv := &server{
		ScyllaSession: session,
		RedisClient:   redisClient,
		NsqProducer:   nsqProducer,
	}

	Logger.WithFields(map[string]interface{}{
		"config": config,
		"time":   time.Now().UnixNano(),
	}).Errorf("start msg")

	grpcServer := runGrpcServer(shutdownObserver, sv)

	// health check
	go runHealth()

	// nsq consumer
	// snowflakeNode, graceful shutdown 使用
	snowflakeNode, err := newSnowFlake()
	if err != nil {
		return
	}
	// 建立空白設定檔。
	ConsumerConfig := nsq.NewConfig()
	// 設置重連時間
	ConsumerConfig.LookupdPollInterval = time.Second * 2
	consumer, _ := nsq.NewConsumer("COCONUT_UPDATE_POINT", "coconut", ConsumerConfig)
	consumer.AddConcurrentHandlers(TestNSQConsumer(), config.NsqConsumerWorkers)
	err = consumer.ConnectToNSQLookupd(config.NsqLookupdAddr)
	if err != nil {
		return
	}

	/** 監聽信號
	  SIGHUP 終端控制進程結束(終端連接斷開)
	  SIGINT 用戶發送INTR字符(Ctrl+C)觸發
	  SIGQUIT 用戶發送QUIT字符(Ctrl+/)觸發
	  SIGTERM 結束程序(可以被捕獲、阻塞或忽略)
	**/
	signal.Notify(shutdownObserver, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	//阻塞直到有信號傳入
	s := <-shutdownObserver
	Logger.Debugf(`Receive signal: %s`, s)

	// 優雅停止GRPC服務
	grpcServer.GracefulStop()

	var (
		wg                  sync.WaitGroup
		stopOnSignalExitNCL = NsqConsumerList{} // 收集停止訊號時停止nsq consumer 的列表
	)

	stopOnSignalExitNCL.Set(snowflakeNode.Generate().Int64(), consumer)
	// 停止部分Nsq Consumer避免有訊息進來
	Logger.Debugf("停止部分Nsq Consumer避免有訊息進來...")

	err = stopOnSignalExitNCL.Each(func(c *nsq.Consumer) error {
		wg.Add(1)
		go func(c *nsq.Consumer, wg *sync.WaitGroup) {
			// 停止訊號會等待正在處理的訊息做完才結束
			c.Stop()
			<-c.StopChan
			wg.Done()
		}(c, &wg)

		return nil
	})

	if err != nil {
		Logger.Errorf("stopOnSignalExitNCL.Each err: %s", err.Error())
	}

	wg.Wait()
	Logger.Debugf("停止 NSQ Consumer完成...")

	// 避免有執行完的動作，休息一下再退出
	for t := 10; t > 0; t-- {
		log.Printf("休息%d秒後準備退出", t)
		time.Sleep(time.Second * 1)
	}
}

// newRedisConnection ...
func newRedisConnection(addr string, maxIdle int, maxConn int, db int) (client *redis.Client, err error) {
	client = redis.NewClient(&redis.Options{
		Addr:       addr,
		DB:         db,
		MaxRetries: 5,
	})

	_, err = client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return
}

// newCQLSession ...
func newCQLSession() (session *gocql.Session, err error) {
	cluster := gocql.NewCluster(config.CqlClusterAddrs...)
	cluster.Port = config.CqlPort
	cluster.Keyspace = config.CqlKeyspace
	cluster.Consistency = ConsistencyStrToConsistency(config.CqlConsistency)
	cluster.Timeout = time.Duration(int64(config.CqlTimeout)) * time.Second
	cluster.ConnectTimeout = time.Duration(int64(config.CqlConnectTimeout)) * time.Second
	cluster.MaxPreparedStmts = config.CqlMaxPreparedStmts
	cluster.NumConns = config.CqlNumConns
	session, err = cluster.CreateSession()

	return
}

func runGrpcServer(c chan<- os.Signal, sr *server) *grpc.Server {
	// 監聽指定埠口，這樣服務才能在該埠口執行。
	apiListener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		panic(err)
	}

	gs := grpc.NewServer(grpc.MaxRecvMsgSize(1024 * 1024 * 8))
	coconut.RegisterCoconutServer(gs, sr)
	// 在 gRPC 伺服器上註冊反射服務。
	reflection.Register(gs)

	go func(gs *grpc.Server, c chan<- os.Signal) {
		err := gs.Serve(apiListener)

		if err != nil {
			c <- syscall.SIGINT
		}
	}(gs, c)

	return gs
}

func runHealth() {
	started := time.Now()
	connSelf, err := grpc.Dial("localhost"+config.Listen, grpc.WithInsecure())

	if err != nil {
		Logger.WithFields(map[string]interface{}{"error": err}).Errorf("連線失敗")
	}

	coco := coconut.NewCoconutClient(connSelf)

	Logger.Debugf("Healthy API is Running at port: %s", config.HealthPort)
	http.HandleFunc(config.HealthPath, func(w http.ResponseWriter, r *http.Request) {
		// 確認gRPC服務有通
		ping, err := coco.Ping(context.Background(), &coconut.PingRequest{})
		if err == nil {
			w.WriteHeader(200)
			data := fmt.Sprintf("Already run: %v", time.Since(started))
			if _, errw := w.Write([]byte(data)); errw != nil {
				Logger.WithFields(map[string]interface{}{"error": errw}).Errorf("runHealth")
			}
		} else {
			w.WriteHeader(500)
			Logger.Errorf("Service Not Ready yet, ping: %#v, err: %s", ping, err.Error())
		}
	})

	logrus.Fatal(http.ListenAndServe(":"+config.HealthPort, nil))
}

func newNSQProducer(addr string) (r *nsq.Producer, err error) {
	NSQconfig := nsq.NewConfig()
	err = NSQconfig.Set("max_in_flight", config.NsqdMaxInFlight)
	if err != nil {
		Logger.WithFields(map[string]interface{}{
			"err":           err,
			"now_time":      time.Now().UnixNano(),
			"max_in_flight": config.NsqdMaxInFlight,
		}).Debugf("newNSQProducer")

		return
	}

	r, err = nsq.NewProducer(addr, NSQconfig)
	if err != nil {
		return
	}

	err = r.Ping()

	if err != nil {
		return
	}

	return r, err
}

// newSnowFlake ...
func newSnowFlake() (node *snowflake.Node, err error) {
	// nodeID for 測試用 直接使用 1
	node, err = snowflake.NewNode(int64(1))
	if err != nil {
		return
	}
	return
}
