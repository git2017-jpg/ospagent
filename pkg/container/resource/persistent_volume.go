package resource

import (
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/websocket"
	"k8s.io/api/core/v1"
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

//func (p *PersistentVolume) List(reuqestParams interface{}) *utils.Response {
//
//}
