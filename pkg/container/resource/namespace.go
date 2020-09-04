package resource

import (
	"encoding/json"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
	"github.com/openspacee/ospagent/pkg/utils/code"
	"github.com/openspacee/ospagent/pkg/websocket"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"strings"
)

type Namespace struct {
	*kubernetes.KubeClient
	websocket.SendResponse
	watch *WatchResource
}

func NewNamespace(
	kubeClient *kubernetes.KubeClient,
	sendResponse websocket.SendResponse,
	watch *WatchResource) *Namespace {

	ns := &Namespace{
		KubeClient:   kubeClient,
		SendResponse: sendResponse,
		watch:        watch,
	}
	ns.DoWatch()
	return ns
}

func (n *Namespace) DoWatch() {
	nsInformer := n.KubeClient.NamespaceInformer().Informer()
	nsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    n.watch.WatchAdd(utils.WatchNamespace),
		UpdateFunc: n.watch.WatchUpdate(utils.WatchNamespace),
		DeleteFunc: n.watch.WatchDelete(utils.WatchNamespace),
	})
}

type NsQueryParams struct {
	Name   string `json:"name"`
	Output string `json:"output"`
}

type BuildNamespace struct {
	UID     string      `json:"uid"`
	Name    string      `json:"name"`
	Created metav1.Time `json:"created"`
	Status  string      `json:"status"`
}

func (n *Namespace) ToBuildNamespace(ns *v1.Namespace) *BuildNamespace {
	if ns == nil {
		return nil
	}
	return &BuildNamespace{
		UID:     string(ns.UID),
		Name:    ns.Name,
		Created: ns.CreationTimestamp,
		Status:  string(ns.Status.Phase),
	}
}

func (n *Namespace) List(requestParams interface{}) *utils.Response {
	queryParams := &NsQueryParams{}
	json.Unmarshal(requestParams.([]byte), queryParams)
	nsList, err := n.KubeClient.InformerRegistry.NamespaceInformer().Lister().List(labels.Everything())
	if err != nil {
		return &utils.Response{
			Code: code.ListError,
			Msg:  err.Error(),
		}
	}
	var nsRes []*BuildNamespace
	for _, ns := range nsList {
		if queryParams.Name == "" || strings.Contains(ns.ObjectMeta.Name, queryParams.Name) {
			nsRes = append(nsRes, n.ToBuildNamespace(ns))
		}
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: nsRes}
}
