package resource

import (
	"fmt"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
	"github.com/openspacee/ospagent/pkg/utils/code"
	"github.com/openspacee/ospagent/pkg/websocket"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"math"
	"time"
)

type Node struct {
	*kubernetes.KubeClient
	websocket.SendResponse
}

type BuildNode struct {
	UID     string `json:"uid"`
	Name    string `json:"name"`
	Taints  int    `json:"taints"`
	Roles   string `json:"roles"`
	Version string `json:"version"`
	Age     string `json:"age"`
	Status  string `json:"status"`
}

func (n *Node) ToBuildNode(node *v1.Node) *BuildNode {
	if node == nil {
		return nil
	}
	nodeData := &BuildNode{
		UID:     string(node.UID),
		Name:    node.Name,
		Taints:  len(node.Spec.Taints),
		Version: node.Status.NodeInfo.KubeletVersion,
	}
	dur := time.Now().Sub(node.CreationTimestamp.Time)
	nodeData.Age = fmt.Sprintf("%vd", math.Floor(dur.Hours()/24))

	for _, c := range node.Status.Conditions {
		if c.Type == "Ready" && c.Status == v1.ConditionTrue {
			nodeData.Status = "Ready"
		} else {
			nodeData.Status = "NotReady"
		}
	}

	return nodeData
}

func NewNode(kubeClient *kubernetes.KubeClient, sendResponse websocket.SendResponse) *Node {
	return &Node{
		KubeClient:   kubeClient,
		SendResponse: sendResponse,
	}
}

func (n *Node) List(requestParams interface{}) *utils.Response {
	nodeList, err := n.KubeClient.InformerRegistry.NodeInformer().Lister().List(labels.Everything())
	if err != nil {
		return &utils.Response{
			Code: code.ListError,
			Msg:  err.Error(),
		}
	}
	var nodeResource []*BuildNode
	for _, node := range nodeList {
		nodeResource = append(nodeResource, n.ToBuildNode(node))
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: nodeResource}
}
