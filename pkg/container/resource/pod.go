package resource

import (
	"encoding/json"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"strings"
)

type Pod struct {
	*kubernetes.KubeClient
}

func NewPod(kubeClient *kubernetes.KubeClient) *Pod {
	return &Pod{
		KubeClient: kubeClient,
	}
}

type listRequestParams struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func (p *Pod) List(requestParams interface{}) *utils.Response {
	listParams := &listRequestParams{}
	json.Unmarshal(requestParams.([]byte), listParams)
	podList, err := p.KubeClient.PodLister().Pods(listParams.Namespace).List(labels.Everything())
	if err != nil {
		return &utils.Response{
			Code: "ListError",
			Msg:  err.Error(),
		}
	}
	var podRes []*v1.Pod
	if listParams.Name != "" {
		for _, pod := range podList {
			if strings.Contains(pod.ObjectMeta.Name, listParams.Name) {
				podRes = append(podRes, pod)
			}
		}
	} else {
		podRes = podList
	}
	return &utils.Response{Code: "Success", Msg: "Success", Data: podRes}
}
