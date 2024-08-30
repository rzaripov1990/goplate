package reqresp

import "fmt"

type (
	Base struct {
		Success bool `json:"success"`
	}

	Data[T any] struct {
		Base
		Data T `json:"data,omitempty"`
	}

	Error struct {
		Base
		Msg        *string `json:"msg,omitempty"`
		MsgType    *string `json:"msgType,omitempty"`
		LogError   error   `json:"-"`
		StatusCode int     `json:"-"`
	}
)

func NewData[T any](value T) Data[T] {
	return Data[T]{
		Base: Base{
			Success: true,
		},
		Data: value,
	}
}

func NewError(statusCode int, err error, message any, msgType string) *Error {
	var m string
	switch t := message.(type) {
	case string:
		m = t
	case []byte:
		m = string(t)
	case error:
		m = t.Error()
	default:
		m = fmt.Sprintf("%v", t)
	}

	return &Error{
		Base: Base{
			Success: false,
		},
		Msg:        &m,
		MsgType:    &msgType,
		LogError:   err,
		StatusCode: statusCode,
	}
}

func (e Error) Error() string {
	switch {
	case e.Msg != nil && e.MsgType != nil:
		return fmt.Sprintf("%s [%s]", *e.Msg, *e.MsgType)
	case e.Msg != nil && e.MsgType == nil:
		return *e.Msg
	default:
		return ""
	}
}
