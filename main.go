package main

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/syhlion/requestwork.v2"
	"github.com/urfave/cli"

	coconutLog "github.com/evelynocean/coconut/lib/log"
)

var (
	config   *Config
	Logger   *coconutLog.Logger
	cmdStart = cli.Command{
		Name:   "start",
		Usage:  "start",
		Action: start,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name: "env-file,e",
			},
			cli.BoolFlag{
				Name:  "debug,d",
				Usage: "open debug mode",
			},
		},
	}
)

func init() {
	Logger = coconutLog.New()
}

func main() {
	// cli
	cmdCli := cli.NewApp()

	cmdCli.Compiled = time.Now()
	cmdCli.Commands = []cli.Command{
		cmdStart,
	}
	err := cmdCli.Run(os.Args)
	if err != nil {
		Logger.WithFields(map[string]interface{}{
			"err:": err.Error(),
		}).Errorf("service run")
	}
}

func envInit(c *cli.Context) {
	/*env init*/
	config = &Config{}
	config.Worker = requestwork.New(1000)

	if c.String("env-file") != "" {
		envfile := c.String("env-file")
		err := godotenv.Load(envfile)
		if err != nil {
			Logger.WithFields(map[string]interface{}{
				"err:": err.Error(),
			}).Errorf("service run")
			return
		}
	}

	var err error

	defer func() {
		if err != nil {
			Logger.WithFields(map[string]interface{}{
				"err:": err.Error(),
			}).Errorf("service init")
		}
	}()
	config.Listen = os.Getenv("API_LISTEN")
	if config.Listen == "" {
		err = errors.New(`env API_LISTEN empty`)
		return
	}

	config.RedisAddr = os.Getenv("REDIS_ADDR")
	if config.RedisAddr == "" {
		err = errors.New(`env REDIS_ADDR empty`)
		return
	}

	config.RedisMaxIdle, err = strconv.Atoi(os.Getenv("REDIS_MAX_IDLE"))
	if err != nil {
		err = errors.New(`env REDIS_MAX_IDLE empty`)
		return
	}

	config.RedisMaxConn, err = strconv.Atoi(os.Getenv("REDIS_MAX_CONN"))
	if err != nil {
		err = errors.New(`env REDIS_MAX_CONN empty`)
		return
	}

	config.RedisNo, err = strconv.Atoi(os.Getenv("REDIS_DB_NO"))
	if err != nil {
		err = errors.New(`env REDIS_DB_NO empty`)
		return
	}

	cqlAddrs := os.Getenv("CQL_CLUSTER_ADDR")
	if cqlAddrs == "" {
		err = errors.New(`env CQL_CLUSTER_ADDR empty`)
		return
	}
	addrs := make([]string, 0)
	err = json.Unmarshal([]byte(cqlAddrs), &addrs)
	if err != nil {
		return
	}
	config.CqlClusterAddrs = addrs

	config.CqlKeyspace = os.Getenv("CQL_KEYSPACE")
	if config.CqlKeyspace == "" {
		err = errors.New(`env CQL_KEYSPACE empty`)
		return
	}

	config.CqlConsistency = os.Getenv("CQL_ORDER_READ_CONSISTENCY")
	if config.CqlConsistency == "" {
		err = errors.New(`env CQL_ORDER_READ_CONSISTENCY empty`)
		return
	}

	config.CqlPort, err = strconv.Atoi(os.Getenv("CQL_PORT"))
	if err != nil {
		err = errors.New(`env CQL_PORT empty`)
		return
	}

	config.CqlTimeout, err = strconv.Atoi(os.Getenv("CQL_TIMEOUT"))
	if err != nil {
		err = errors.New(`env CQL_TIMEOUT empty`)
		return
	}

	config.CqlConnectTimeout, err = strconv.Atoi(os.Getenv("CQL_CONNECTION_TIMEOUT"))
	if err != nil {
		err = errors.New(`env CQL_CONNECTION_TIMEOUT empty`)
		return
	}

	config.CqlMaxPreparedStmts, err = strconv.Atoi(os.Getenv("CQL_MAX_PREPARED_STMTS"))
	if err != nil {
		err = errors.New(`env CQL_MAX_PREPARED_STMTS empty`)
		return
	}

	config.CqlNumConns, err = strconv.Atoi(os.Getenv("CQL_NUM_CONNS"))
	if err != nil {
		err = errors.New(`env CQL_NUM_CONNS empty`)
		return
	}

}
