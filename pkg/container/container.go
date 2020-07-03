package container

import (
	"encoding/json"
	"fmt"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
	"k8s.io/klog"
)

type Container struct {
	KubeClient   *kubernetes.KubeClient
	RequestChan  chan *utils.Request
	ResponseChan chan *utils.TResponse
	*ResourceActions
}

func NewContainer(
	kubeClient *kubernetes.KubeClient,
	requestChan chan *utils.Request,
	responseChan chan *utils.TResponse) *Container {

	resourceActions := NewResourceActions(kubeClient)
	return &Container{
		KubeClient:      kubeClient,
		RequestChan:     requestChan,
		ResponseChan:    responseChan,
		ResourceActions: resourceActions,
	}
}

func (c *Container) Run() {
	for {
		select {
		case req, ok := <-c.RequestChan:
			if ok {
				go c.handleRequest(req)
			}
		}
	}
}

func (c *Container) handleRequest(request *utils.Request) {
	resp := c.doRequest(request)
	tResp := &utils.TResponse{RequestId: request.RequestId, Data: resp}
	c.ResponseChan <- tResp
}

func (c *Container) doRequest(request *utils.Request) (resp *utils.Response) {
	defer func() {
		if err := recover(); err != nil {
			klog.Error("do request error: ", err)
			msg := fmt.Sprintf("%s", err)
			resp = &utils.Response{Code: "UnknownError", Msg: msg}
		}
	}()
	resource := request.Resource
	action := request.Action
	params := request.Params
	handler := c.GetRequestHandler(resource, action)
	resp = &utils.Response{}
	if handler == nil {
		msg := fmt.Sprintf("resource %s action %s not found", resource, action)
		klog.Error(msg)
		resp = &utils.Response{Code: "ActionError", Msg: msg}
	} else {
		jsonParams, _ := json.Marshal(params)
		resp = handler(jsonParams)
	}
	return
}
