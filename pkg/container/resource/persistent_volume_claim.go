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

type PersistentVolumeClaim struct {
	*kubernetes.KubeClient
	websocket.SendResponse
	*DynamicResource
}

type BuildPersistentVolumeClaim struct {
	Name          string                           `json:"name"`
	Namespace     string                           `json:"namespace"`
	Status        string                           `json:"status"`
	StorageClass  string                           `json:"storage_class"`
	Capacity      string                           `json:"capacity"`
	AccessModes   []v1.PersistentVolumeAccessMode  `json:"access_modes"`
	CreateTime    string                           `json:"create_time"`
	ReclaimPolicy v1.PersistentVolumeReclaimPolicy `json:"reclaim_policy"`
}

func NewPersistentVolumeClaim(kubeClient *kubernetes.KubeClient, sendResponse websocket.SendResponse) *PersistentVolumeClaim {
	return &PersistentVolumeClaim{
		KubeClient:   kubeClient,
		SendResponse: sendResponse,
		DynamicResource: NewDynamicResource(kubeClient, &schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "persistentVolumeClaim",
		}),
	}
}

func (p *PersistentVolumeClaim) ToBuildPersistentVolumeClaim(pvc *v1.PersistentVolumeClaim) *BuildPersistentVolumeClaim {
	if pvc == nil {
		return nil
	}

	pvcData := &BuildPersistentVolumeClaim{
		Name:        pvc.Name,
		Status:      string(pvc.Status.Phase),
		AccessModes: pvc.Spec.AccessModes,
		CreateTime:  fmt.Sprint(pvc.CreationTimestamp),
		Namespace:   pvc.Namespace,
	}

	return pvcData
}

func (p *PersistentVolumeClaim) List(requestParams interface{}) *utils.Response {
	persistentVolumeClaimList, err := p.KubeClient.InformerRegistry.PersistentVolumeClaimInformer().Lister().List(labels.Everything())
	if err != nil {
		return &utils.Response{
			Code: code.ListError,
			Msg:  err.Error(),
		}
	}
	var persistentVolumeClaimResource []*BuildPersistentVolumeClaim
	for _, pvc := range persistentVolumeClaimList {
		persistentVolumeClaimResource = append(persistentVolumeClaimResource, p.ToBuildPersistentVolumeClaim(pvc))
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: persistentVolumeClaimResource}
}

func (p *PersistentVolumeClaim) Get(requestParams interface{}) *utils.Response {
	queryParams := &ConfigMapQueryParams{}
	json.Unmarshal(requestParams.([]byte), queryParams)
	if queryParams.Name == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Name is blank"}
	}
	if queryParams.Namespace == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Namespace is blank"}
	}
	pvc, err := p.KubeClient.InformerRegistry.PersistentVolumeClaimInformer().Lister().PersistentVolumeClaims(queryParams.Namespace).Get(queryParams.Name)
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
		d, e := runtime.Encode(encoder, pvc)
		if e != nil {
			klog.Error(e)
			return &utils.Response{Code: code.EncodeError, Msg: e.Error()}
		}
		klog.Info(d)
		return &utils.Response{Code: code.Success, Msg: "Success", Data: string(d)}
	}

	return &utils.Response{Code: code.Success, Msg: "Success", Data: pvc}
}
