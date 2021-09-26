package main

import (
	"github.com/syhlion/requestwork.v2"
)

type Config struct {
	Listen              string
	Worker              *requestwork.Worker
	RedisAddr           string
	RedisMaxIdle        int
	RedisMaxConn        int
	RedisNo             int
	CqlClusterAddrs     []string
	CqlKeyspace         string
	CqlPort             int
	CqlConsistency      string
	CqlTimeout          int
	CqlConnectTimeout   int
	CqlMaxPreparedStmts int
	CqlNumConns         int
	Environment         string
	HealthPort          string
	HealthPath          string
	NsqdAddr            string
	NsqdMaxInFlight     int
	NsqLookupdAddr      string
	NsqConsumerWorkers  int
}
