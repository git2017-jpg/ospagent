package resource

import (
	"encoding/json"
	"fmt"
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

type BuildContainer struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Restarts int32  `json:"restarts"`
	Ready    bool   `json:"ready"`
}

type BuildPod struct {
	Name           string            `json:"name"`
	Namespace      string            `json:"namespace"`
	Containers     []*BuildContainer `json:"containers"`
	InitContainers []*BuildContainer `json:"init_containers"`
	Controlled     string            `json:"controlled"`
	Qos            string            `json:"qos"`
	Created        string            `json:"created"`
	Status         string            `json:"status"`
	Ip             string            `json:"ip"`
	NodeName       string            `json:"node_name"`
}

func (p *Pod) ToBuildContainer(statuses []v1.ContainerStatus, container *v1.Container) *BuildContainer {
	bc := &BuildContainer{
		Name: container.Name,
	}
	for _, s := range statuses {
		if s.Name == container.Name {
			bc.Restarts = s.RestartCount
			if s.State.Running != nil {
				bc.Status = "running"
			} else if s.State.Terminated != nil {
				bc.Status = "terminated"
			} else if s.State.Waiting != nil {
				bc.Status = "waiting"
			}
			bc.Ready = s.Ready
			break
		}
	}
	return bc
}

func (p *Pod) ToBuildPod(pod *v1.Pod) *BuildPod {
	if pod == nil {
		return nil
	}
	var containers []*BuildContainer
	for _, container := range pod.Spec.Containers {
		bc := p.ToBuildContainer(pod.Status.ContainerStatuses, &container)
		containers = append(containers, bc)
	}
	var initContainers []*BuildContainer
	for _, container := range pod.Spec.InitContainers {
		bc := p.ToBuildContainer(pod.Status.InitContainerStatuses, &container)
		initContainers = append(initContainers, bc)
	}
	var controlled = ""
	if len(pod.ObjectMeta.OwnerReferences) > 0 {
		controlled = pod.ObjectMeta.OwnerReferences[0].Kind
	}
	return &BuildPod{
		Name:           pod.ObjectMeta.Name,
		Namespace:      pod.ObjectMeta.Namespace,
		Containers:     containers,
		InitContainers: initContainers,
		Controlled:     controlled,
		Qos:            string(pod.Status.QOSClass),
		Status:         string(pod.Status.Phase),
		Ip:             pod.Status.PodIP,
		Created:        fmt.Sprint(pod.ObjectMeta.CreationTimestamp),
		NodeName:       pod.Spec.NodeName,
	}
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
	var podRes []*BuildPod
	for _, pod := range podList {
		if listParams.Name == "" || strings.Contains(pod.ObjectMeta.Name, listParams.Name) {
			podRes = append(podRes, p.ToBuildPod(pod))
		}
	}
	return &utils.Response{Code: "Success", Msg: "Success", Data: podRes}
}
