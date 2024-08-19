package reqresp

type (
	Base struct {
		Success bool `json:"success"`
	}

	Data[T any] struct {
		Base
		Data T `json:"data,omitempty"`
	}

	Error[E, T string] struct {
		Base
		Msg     E `json:"msg,omitempty"`
		MsgType T `json:"msgType,omitempty"`
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

func NewError[E, T string](erro_ E, typ_ T) Error[E, T] {
	return Error[E, T]{
		Base: Base{
			Success: false,
		},
		Msg:     erro_,
		MsgType: typ_,
	}
}
