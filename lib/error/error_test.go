package coconutError

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestError(t *testing.T) {
	e := ErrServer
	e.ExtraInfo = map[string]interface{}{
		"aa": 1,
	}

	testError := ParseError(e, errors.New("測試錯誤格式"))
	str, err := json.Marshal(testError)
	if err != nil {
		t.Errorf("json.Marshal error: %s", err)
	}
	fmt.Println(string(str))
}
