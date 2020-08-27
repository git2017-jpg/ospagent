package resource

import (
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/websocket"
	"k8s.io/api/core/v1"
	"github.com/openspacee/ospagent/pkg/utils"
	"k8s.io/apimachinery/pkg/labels"
	"github.com/openspacee/ospagent/pkg/utils/code"
)

type PersistentVolume struct {
	*kubernetes.KubeClient
	websocket.SendResponse
}

type BuildPersistentVolume struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Claim  string `json:"claim"`
}

func (p *PersistentVolume) ToBuildPersistentVolume(pv *v1.PersistentVolume) *BuildPersistentVolume {
	if pv == nil {
		return nil
	}
	pvData := &BuildPersistentVolume{
		Name:   pv.Name,
		Status: "",
	}

	return pvData
}

func (p *PersistentVolume) List(requestParams interface{}) *utils.Response {
	persistentVolumeList, err := p.KubeClient.InformerRegistry.PersistentVolumeInformer().Lister().List(labels.Everything())
	if err != nil {
		return &utils.Response{
			Code: code.ListError,
			Msg: err.Error(),
		}
	}
	var persistentVolumeResource []*BuildPersistentVolume
	for _, pv := range persistentVolumeList {
		persistentVolumeResource = append(persistentVolumeResource, p.ToBuildPersistentVolume(pv))
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: persistentVolumeResource}
}
