package utils

import (
	"encoding/json"
	"github.com/openspacee/ospagent/pkg/utils/code"
)

type Response struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type TResponse struct {
	RequestId string    `json:"request_id"`
	Data      *Response `json:"data"`
}

func (resp *TResponse) Serializer() ([]byte, error) {
	return json.Marshal(resp)
}

func (r *Response) IsSuccess() bool {
	return r.Code == code.Success
}
