package resource

import (
	"encoding/json"
	"fmt"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
	"github.com/openspacee/ospagent/pkg/utils/code"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

//type Secret struct {
//	*kubernetes.KubeClient
//	websocket.SendResponse
//	*DynamicResource
//}

type Secret struct {
	watch *WatchResource
	*DynamicResource
}

type BuildSecret struct {
	Name       string            `json:"name"`
	NameSpace  string            `json:"namespace"`
	Keys       []string          `json:"keys"`
	Labels     map[string]string `json:"labels"`
	CreateTime string            `json:"create_time"`
	Type       v1.SecretType     `json:"type"`
	Data       map[string][]byte `json:"data"`
}

type SecretQueryParams struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Output    string `json:"output"`
}

func (s *Secret) ToBuildSecret(se *v1.Secret) *BuildSecret {
	if se == nil {
		return nil
	}

	sData := &BuildSecret{
		Name:       se.Name,
		NameSpace:  se.Namespace,
		Labels:     se.Labels,
		Type:       se.Type,
		CreateTime: fmt.Sprint(se.CreationTimestamp),
		Data:       se.Data,
	}
	return sData
}

func NewSecret(kubeClient *kubernetes.KubeClient, watch *WatchResource) *Secret {
	s := &Secret{
		watch: watch,
		DynamicResource: NewDynamicResource(kubeClient, &schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "secrets",
		}),
	}
	s.DoWatch()
	return s
}

func (s *Secret) DoWatch() {
	informer := s.KubeClient.SecretInformer().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.watch.WatchAdd(utils.WatchService),
		UpdateFunc: s.watch.WatchUpdate(utils.WatchService),
		DeleteFunc: s.watch.WatchDelete(utils.WatchService),
	})
}

func (s *Secret) List(requestParams interface{}) *utils.Response {
	secretList, err := s.KubeClient.InformerRegistry.SecretInformer().Lister().List(labels.Everything())
	if err != nil {
		return &utils.Response{
			Code: code.ListError,
			Msg:  err.Error(),
		}
	}
	var secretResource []*BuildSecret
	for _, cm := range secretList {
		secretResource = append(secretResource, s.ToBuildSecret(cm))
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: secretResource}
}

func (s *Secret) Get(requestParams interface{}) *utils.Response {
	queryParams := &SecretQueryParams{}
	json.Unmarshal(requestParams.([]byte), queryParams)
	if queryParams.Name == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Name is blank"}
	}
	if queryParams.Namespace == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Namespace is blank"}
	}
	secret, err := s.KubeClient.SecretInformer().Lister().Secrets(queryParams.Namespace).Get(queryParams.Name)
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
		d, e := runtime.Encode(encoder, secret)
		if e != nil {
			klog.Error(e)
			return &utils.Response{Code: code.EncodeError, Msg: e.Error()}
		}
		klog.Info(d)
		return &utils.Response{Code: code.Success, Msg: "Success", Data: string(d)}
	}
	fmt.Println(secret)
	return &utils.Response{Code: code.Success, Msg: "Success", Data: secret}
}
