package resource

import (
	"fmt"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
	"github.com/openspacee/ospagent/pkg/utils/code"
	"github.com/openspacee/ospagent/pkg/websocket"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type ConfigMap struct {
	*kubernetes.KubeClient
	websocket.SendResponse
}

type BuildConfigMap struct {
	Name       string            `json:"name"`
	NameSpace  string            `json:"namespace"`
	Keys       []string          `json:"keys"`
	Labels     map[string]string `json:"labels"`
	CreateTime string            `json:"create_time"`
	Data       map[string]string `json:"data"`
}

func (c *ConfigMap) ToBuildConfigMap(cm *v1.ConfigMap) *BuildConfigMap {
	if cm == nil {
		return nil
	}
	cmData := &BuildConfigMap{
		Name:       cm.Name,
		NameSpace:  cm.Namespace,
		Labels:     cm.Labels,
		CreateTime: fmt.Sprint(cm.CreationTimestamp),
		Data:       cm.Data,
	}

	keys := make([]string, 0, len(cm.Data))
	for k, _ := range cm.Data {
		keys = append(keys, k)
	}
	cmData.Keys = keys
	return cmData
}

func NewConfigMap(kubeClient *kubernetes.KubeClient, sendResponse websocket.SendResponse) *ConfigMap {
	return &ConfigMap{
		KubeClient:   kubeClient,
		SendResponse: sendResponse,
	}
}

func (c *ConfigMap) List(requestParams interface{}) *utils.Response {
	configMapList, err := c.KubeClient.InformerRegistry.ConfigMapInformer().Lister().List(labels.Everything())
	if err != nil {
		return &utils.Response{
			Code: code.ListError,
			Msg:  err.Error(),
		}
	}
	var configMapResource []*BuildConfigMap
	for _, cm := range configMapList {
		configMapResource = append(configMapResource, c.ToBuildConfigMap(cm))
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: configMapResource}
}
