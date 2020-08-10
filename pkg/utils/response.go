package utils

import (
	"encoding/json"
	"github.com/openspacee/ospagent/pkg/utils/code"
)

const (
	RequestType = "request"
	WatchType   = "watch"
	ExecType    = "exec"
	LogType     = "log"

	AddEvent    = "add"
	UpdateEvent = "update"
	DeleteEvent = "delete"

	WatchPod       = "pods"
	WatchNamespace = "namespace"
)

type Response struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type WatchResponse struct {
	Event    string      `json:"event"`
	Obj      string      `json:"obj"`
	Resource interface{} `json:"resource"`
}

type TResponse struct {
	ResType   string      `json:"res_type"`
	RequestId string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

func (resp *TResponse) Serializer() ([]byte, error) {
	return json.Marshal(resp)
}

func (r *Response) IsSuccess() bool {
	return r.Code == code.Success
}
