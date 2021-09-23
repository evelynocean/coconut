package coconut_model

import (
	"github.com/syhlion/gocql"
)

// Init Init
func Init(cql *gocql.Session) {
	LimitSQL = &RealModel{
		cqlDB: cql,
	}
}

// InitMock InitMock
func InitMock() {
	LimitSQL = &MockModel{}
}
