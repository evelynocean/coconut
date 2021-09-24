package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
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

	if config.Environment == "local" {
		coconut_model.InitMock()
	} else {
		coconut_model.Init(session)
	}

	sv := &server{
		ScyllaSession: session,
		RedisClient:   redisClient,
	}

	startStr, err := json.Marshal(config)
	fmt.Println("start msg:", string(startStr))
	Logger.WithFields(map[string]interface{}{
		"config": config,
		"time":   time.Now().UnixNano(),
	}).Errorf("start msg")

	grpcServer := runGrpcServer(shutdownObserver, sv)

	// health check
	go runHealth()

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
