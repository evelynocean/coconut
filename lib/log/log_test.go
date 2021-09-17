package coconutLog

import (
	"errors"
	"testing"
)

func TestLog(t *testing.T) {
	logger := New()
	logger.Debugf("前天星期%s", "三")
	logger.WithFields(map[string]interface{}{"昨天": "9/9"}).Debugf("星期%s", "四")
	logger.WithFields(map[string]interface{}{"今天": "9/10"}).Errorf("星期%s", "五")
	logger.WithError(errors.New("後天")).Errorf("星期%s", "六")
	logger.Errorf("最後是 星期%s", "天")
}
