package rpc

type RpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewRpcError(code int, message string, data interface{}) *RpcError {
	return &RpcError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func NewParsingError(message string) *RpcError {
	return NewRpcError(-32700, message, nil)
}

func NewInvalidRequestError(message string) *RpcError {
	return NewRpcError(-32600, message, nil)
}

func NewMethodNotFoundError() *RpcError {
	return NewRpcError(-32601, "method not found", nil)
}

func NewInvalidParamsError(message string) *RpcError {
	return NewRpcError(-32602, message, nil)
}

func NewInternalError(message string) *RpcError {
	return NewRpcError(-32603, message, nil)
}

func (e *RpcError) Error() string {
	return e.Message
}
