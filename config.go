package main

import (
	"github.com/syhlion/requestwork.v2"
)

type Config struct {
	Listen       string
	Worker       *requestwork.Worker
	RedisAddr    string
	RedisMaxIdle int
	RedisMaxConn int
	RedisNo      int
	// MysqlSlow     time.Duration
	// MasterAddr    string
	// MasterMaxIdle int
	// MasterMaxConn int
	// SlaveAddr     string
	// SlaveMaxIdle  int
	// SlaveMaxConn  int
	CqlClusterAddrs     []string
	CqlKeyspace         string
	CqlPort             int
	CqlConsistency      string
	CqlTimeout          int
	CqlConnectTimeout   int
	CqlMaxPreparedStmts int
	CqlNumConns         int
}
