package resource

import (
	"encoding/json"
	"fmt"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
	"github.com/openspacee/ospagent/pkg/utils/code"
	"github.com/openspacee/ospagent/pkg/websocket"
	hpa "k8s.io/api/autoscaling/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog"
)

type HorizontalPodAutoscaler struct {
	*kubernetes.KubeClient
	websocket.SendResponse
	*DynamicResource
}

type BuildHorizontalPodAutoscaler struct {
	Name      string `json:"name"`
	NameSpace string `json:"namespace"`
}

type HorizontalPodAutoscalerQueryParams struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Output    string `json:"output"`
}

func (h *HorizontalPodAutoscaler) ToBuildHorizontalPodAutoscaler(hpa *hpa.HorizontalPodAutoscaler) *BuildHorizontalPodAutoscaler {
	if hpa == nil {
		return nil
	}

	hpaData := &BuildHorizontalPodAutoscaler{
		Name:      hpa.Name,
		NameSpace: hpa.Namespace,
	}

	return hpaData
}

func NewHorizontalPodAutoscaler(kubeClient *kubernetes.KubeClient, sendResponse websocket.SendResponse) *HorizontalPodAutoscaler {
	return &HorizontalPodAutoscaler{
		KubeClient:   kubeClient,
		SendResponse: sendResponse,
		DynamicResource: NewDynamicResource(kubeClient, &schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "horizontalPodAutoscaler",
		}),
	}
}

func (h *HorizontalPodAutoscaler) List(requestParams interface{}) *utils.Response {
	HpaList, err := h.KubeClient.InformerRegistry.HorizontalPodAutoscalerInformer().Lister().List(labels.Everything())
	if err != nil {
		return &utils.Response{
			Code: code.ListError,
			Msg:  err.Error(),
		}
	}
	var HpaResource []*BuildHorizontalPodAutoscaler
	for _, hp := range HpaList {
		HpaResource = append(HpaResource, h.ToBuildHorizontalPodAutoscaler(hp))
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: HpaResource}
}

func (h *HorizontalPodAutoscaler) Get(requestParams interface{}) *utils.Response {
	queryParams := &HorizontalPodAutoscalerQueryParams{}
	json.Unmarshal(requestParams.([]byte), queryParams)
	if queryParams.Name == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Name is blank"}
	}
	if queryParams.Namespace == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Namespace is blank"}
	}
	Hpa, err := h.KubeClient.HorizontalPodAutoscalerInformer().Lister().HorizontalPodAutoscalers(queryParams.Namespace).Get(queryParams.Name)
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

		encoder := codecs.EncoderForVersion(info.Serializer, h.GroupVersion())
		d, e := runtime.Encode(encoder, Hpa)
		if e != nil {
			klog.Error(e)
			return &utils.Response{Code: code.EncodeError, Msg: e.Error()}
		}
		klog.Info(d)
		return &utils.Response{Code: code.Success, Msg: "Success", Data: string(d)}
	}

	return &utils.Response{Code: code.Success, Msg: "Success", Data: Hpa}
}
