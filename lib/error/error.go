package coconutError

import (
	"fmt"
	"time"

	coconutLog "github.com/evelynocean/coconut/lib/log"
)

// const service = "Coconut"

type Error struct {
	Msg         string                 `json:"msg"`
	Code        int                    `json:"code"`
	ExtraInfo   map[string]interface{} `json:"extrainfo"`
	Time        int64                  `json:"time"`
	Service     string                 `json:"service"`
	OriginError string                 `json:"origin_error"`
}

var (
	Logger *coconutLog.Logger
	// ErrServer ...
	ErrServer = &Error{Code: 9999, Msg: "ERROR_SERVER", ExtraInfo: make(map[string]interface{})}
	// ErrRedis ...
	ErrRedis = &Error{Code: 9998, Msg: "ERROR_REDIS", ExtraInfo: make(map[string]interface{})}
)

func init() {
	Logger = coconutLog.New()
}

// ParseError error 輸出前加工
func ParseError(e *Error, err error) *Error {
	e.Service = "Coconut"
	e.Time = time.Now().Unix()
	e.OriginError = err.Error()

	logMsg := map[string]interface{}{
		"service":      "Coconut",
		"time":         time.Now().Unix(),
		"origin_error": err.Error(),
		"code":         e.Code,
		"msg":          e.Msg,
		"extra_info":   e.ExtraInfo,
	}
	Logger.WithFields(logMsg).Errorf(e.Msg)

	return e
}

func (t Error) Error() string {
	return fmt.Sprintf("[Error %d] %s", t.Code, t.Msg)
}
