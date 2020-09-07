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

type PersistentVolume struct {
	*kubernetes.KubeClient
	websocket.SendResponse
	*DynamicResource
}

type BuildPersistentVolume struct {
	Name          string                           `json:"name"`
	Status        string                           `json:"status"`
	Claim         string                           `json:"claim"`
	StorageClass  string                           `json:"storage_class"`
	Capacity      string                           `json:"capacity"`
	AccessModes   []v1.PersistentVolumeAccessMode  `json:"access_modes"`
	CreateTime    string                           `json:"create_time"`
	ReclaimPolicy v1.PersistentVolumeReclaimPolicy `json:"reclaim_policy"`
}

func NewPersistentVolume(kubeClient *kubernetes.KubeClient, sendResponse websocket.SendResponse) *PersistentVolume {
	return &PersistentVolume{
		KubeClient:   kubeClient,
		SendResponse: sendResponse,
		DynamicResource: NewDynamicResource(kubeClient, &schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "persistentVolume",
		}),
	}
}

func (p *PersistentVolume) ToBuildPersistentVolume(pv *v1.PersistentVolume) *BuildPersistentVolume {
	if pv == nil {
		return nil
	}
	var volumeSize string
	if size, ok := pv.Spec.Capacity["storage"]; !ok {
		volumeSize = ""
	} else {
		volumeSize = size.String()
	}
	pvData := &BuildPersistentVolume{
		Name:          pv.Name,
		Status:        string(pv.Status.Phase),
		StorageClass:  pv.Spec.StorageClassName,
		Capacity:      volumeSize,
		Claim:         pv.Spec.ClaimRef.Name,
		AccessModes:   pv.Spec.AccessModes,
		ReclaimPolicy: pv.Spec.PersistentVolumeReclaimPolicy,
		CreateTime:    fmt.Sprint(pv.CreationTimestamp),
	}

	return pvData
}

func (p *PersistentVolume) List(requestParams interface{}) *utils.Response {
	persistentVolumeList, err := p.KubeClient.InformerRegistry.PersistentVolumeInformer().Lister().List(labels.Everything())
	if err != nil {
		return &utils.Response{
			Code: code.ListError,
			Msg:  err.Error(),
		}
	}
	var persistentVolumeResource []*BuildPersistentVolume
	for _, pv := range persistentVolumeList {
		persistentVolumeResource = append(persistentVolumeResource, p.ToBuildPersistentVolume(pv))
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: persistentVolumeResource}
}

func (p *PersistentVolume) Get(requestParams interface{}) *utils.Response {
	queryParams := &ConfigMapQueryParams{}
	json.Unmarshal(requestParams.([]byte), queryParams)
	if queryParams.Name == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Name is blank"}
	}
	pv, err := p.KubeClient.PersistentVolumeInformer().Lister().Get(queryParams.Name)
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

		encoder := codecs.EncoderForVersion(info.Serializer, p.GroupVersion())
		d, e := runtime.Encode(encoder, pv)
		if e != nil {
			klog.Error(e)
			return &utils.Response{Code: code.EncodeError, Msg: e.Error()}
		}
		klog.Info(d)
		return &utils.Response{Code: code.Success, Msg: "Success", Data: string(d)}
	}

	return &utils.Response{Code: code.Success, Msg: "Success", Data: pv}
}
