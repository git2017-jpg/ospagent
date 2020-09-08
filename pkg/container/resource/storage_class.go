package resource

import (
	"encoding/json"
	"fmt"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
	"github.com/openspacee/ospagent/pkg/utils/code"
	"github.com/openspacee/ospagent/pkg/websocket"
	core "k8s.io/api/core/v1"
	"k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog"
)

type StorageClass struct {
	*kubernetes.KubeClient
	websocket.SendResponse
	*DynamicResource
}

type BuildStorageClass struct {
	Name          string                              `json:"name"`
	CreateTime    string                              `json:"create_time"`
	Provisioner   string                              `json:"provisioner"`
	ReclaimPolicy *core.PersistentVolumeReclaimPolicy `json:"reclaim_policy"`
	Default       string                              `json:"default"`
}

type StorageClassQueryParams struct {
	Name   string `json:"name"`
	Output string `json:"output"`
}

func NewStorageClass(kubeClient *kubernetes.KubeClient, sendResponse websocket.SendResponse) *StorageClass {
	return &StorageClass{
		KubeClient:   kubeClient,
		SendResponse: sendResponse,
		DynamicResource: NewDynamicResource(kubeClient, &schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "storageClass",
		}),
	}
}

func (s *StorageClass) ToBuildStorageClass(sc *v1.StorageClass) *BuildStorageClass {
	if sc == nil {
		return nil
	}

	pvData := &BuildStorageClass{
		Name:          sc.Name,
		CreateTime:    fmt.Sprint(sc.CreationTimestamp),
		Provisioner:   sc.Provisioner,
		ReclaimPolicy: sc.ReclaimPolicy,
	}

	return pvData
}

func (s *StorageClass) List(requestParams interface{}) *utils.Response {
	persistentVolumeList, err := s.KubeClient.InformerRegistry.StorageClassInformer().Lister().List(labels.Everything())
	if err != nil {
		return &utils.Response{
			Code: code.ListError,
			Msg:  err.Error(),
		}
	}
	var persistentVolumeResource []*BuildStorageClass
	for _, pv := range persistentVolumeList {
		persistentVolumeResource = append(persistentVolumeResource, s.ToBuildStorageClass(pv))
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: persistentVolumeResource}
}

func (s *StorageClass) Get(requestParams interface{}) *utils.Response {
	queryParams := &StorageClassQueryParams{}
	json.Unmarshal(requestParams.([]byte), queryParams)
	if queryParams.Name == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Name is blank"}
	}
	sc, err := s.KubeClient.StorageClassInformer().Lister().Get(queryParams.Name)
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

		encoder := codecs.EncoderForVersion(info.Serializer, s.GroupVersion())
		d, e := runtime.Encode(encoder, sc)
		if e != nil {
			klog.Error(e)
			return &utils.Response{Code: code.EncodeError, Msg: e.Error()}
		}
		klog.Info(d)
		return &utils.Response{Code: code.Success, Msg: "Success", Data: string(d)}
	}

	return &utils.Response{Code: code.Success, Msg: "Success", Data: sc}
}
