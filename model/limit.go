package coconut_model

import "github.com/syhlion/gocql"

func GetLimit(cqlDB *gocql.Session) (result map[string]int, err error) {
	var (
		level string
		limit int
	)

	result = make(map[string]int)

	sql := `select level, limit_point from coconut.settings`
	iter := cqlDB.Query(sql).Iter()

	for iter.Scan(&level, &limit) {
		result[level] = limit
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return
}
