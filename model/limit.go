package coconut_model

import "github.com/syhlion/gocql"

type IModel interface {
	GetLimit() (result map[string]int, err error)
}

var LimitSQL IModel

type RealModel struct {
	cqlDB *gocql.Session
}

func (r *RealModel) GetLimit() (result map[string]int, err error) {
	var (
		level string
		limit int
	)

	result = make(map[string]int)

	sql := `select level, limit_point from coconut.settings`
	iter := r.cqlDB.Query(sql).Iter()

	for iter.Scan(&level, &limit) {
		result[level] = limit
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return
}

type MockModel struct{}

func (m *MockModel) GetLimit() (result map[string]int, err error) {
	result = make(map[string]int, 0)
	result["0"] = 11111
	result["1"] = 2222
	result["2"] = 333
	result["3"] = 55

	return result, nil
}
