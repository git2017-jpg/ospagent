package resource

import (
	"encoding/json"
	"fmt"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
	"github.com/openspacee/ospagent/pkg/utils/code"
	"github.com/openspacee/ospagent/pkg/websocket"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog"
)

type ConfigMap struct {
	*kubernetes.KubeClient
	websocket.SendResponse
	*DynamicResource
}

type BuildConfigMap struct {
	Name       string            `json:"name"`
	NameSpace  string            `json:"namespace"`
	Keys       []string          `json:"keys"`
	Labels     map[string]string `json:"labels"`
	CreateTime string            `json:"create_time"`
	Data       map[string]string `json:"data"`
}

type ConfigMapQueryParams struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Output    string `json:"output"`
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
		DynamicResource: NewDynamicResource(kubeClient, &schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "configMap",
		}),
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

func (c *ConfigMap) Get(requestParams interface{}) *utils.Response {
	queryParams := &ConfigMapQueryParams{}
	json.Unmarshal(requestParams.([]byte), queryParams)
	if queryParams.Name == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Name is blank"}
	}
	if queryParams.Namespace == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Namespace is blank"}
	}
	configMap, err := c.KubeClient.ConfigMapInformer().Lister().ConfigMaps(queryParams.Namespace).Get(queryParams.Name)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	if queryParams.Output == "yaml" {
		const mediaType = runtime.ContentTypeYAML
		rscheme := runtime.NewScheme()
		v1.AddToScheme(rscheme)
		codecs := serializer.NewCodecFactory(rscheme)
		info, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), mediaType)
		if !ok {
			return &utils.Response{Code: code.Success, Msg: fmt.Sprintf("unsupported media type %q", mediaType)}
		}

		encoder := codecs.EncoderForVersion(info.Serializer, c.GroupVersion())
		d, e := runtime.Encode(encoder, configMap)
		if e != nil {
			klog.Error(e)
			return &utils.Response{Code: code.EncodeError, Msg: e.Error()}
		}
		klog.Info(d)
		return &utils.Response{Code: code.Success, Msg: "Success", Data: string(d)}
	}

	return &utils.Response{Code: code.Success, Msg: "Success", Data: configMap}
}
