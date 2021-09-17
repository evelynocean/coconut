package main

import (
	"strings"

	"github.com/syhlion/gocql"
)

// ConsistencyStrToConsistency scylla 策略轉換
func ConsistencyStrToConsistency(c string) (cs gocql.Consistency) {
	switch strings.ToUpper(c) {
	case "ALL":
		cs = gocql.All
	case "EACH_QUORUM":
		cs = gocql.EachQuorum
	case "LOCAL_QUORUM":
		cs = gocql.LocalQuorum
	case "LOCAL_ONE":
		cs = gocql.LocalOne
	case "ONE":
		cs = gocql.One
	case "QUORUM":
		cs = gocql.Quorum
	case "TWO":
		cs = gocql.Two
	case "THREE":
		cs = gocql.Three
	}

	return cs
}

func HandlerPanicRecover(err *error) {
	if r := recover(); r != nil {
		Logger.WithError(*err).Errorf("HandlerPanicRecover")
	}
}
