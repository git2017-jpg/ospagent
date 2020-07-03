package container

import (
	"github.com/openspacee/ospagent/pkg/container/resource"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
)

const (
	LIST = "list"
)

type Handler func(interface{}) *utils.Response

type ActionHandler map[string]Handler

type ResourceActions struct {
	KubeClient            *kubernetes.KubeClient
	ResourceActionHandler map[string]ActionHandler
}

func NewResourceActions(kubeClient *kubernetes.KubeClient) *ResourceActions {
	actionHandlers := make(map[string]ActionHandler)

	pod := resource.NewPod(kubeClient)
	podActions := ActionHandler{
		LIST: pod.List,
	}
	actionHandlers["pod"] = podActions
	return &ResourceActions{
		KubeClient:            kubeClient,
		ResourceActionHandler: actionHandlers,
	}
}

func (r *ResourceActions) GetRequestHandler(resource string, action string) Handler {
	return r.ResourceActionHandler[resource][action]
}
