package util

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func WriteResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := Response{
		Code: code,
		Msg:  http.StatusText(code),
		Data: data,
	}
	jsonResponse, _ := json.Marshal(response)
	w.WriteHeader(code)
	w.Write(jsonResponse)
}
