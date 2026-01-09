package dto

// Response 统一响应格式
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) *Response {
	return &Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(code int, msg string) *Response {
	return &Response{
		Code: code,
		Msg:  msg,
	}
}
