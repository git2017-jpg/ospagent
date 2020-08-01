package core

import (
	"github.com/openspacee/ospagent/pkg/config"
	"github.com/openspacee/ospagent/pkg/container"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
	"github.com/openspacee/ospagent/pkg/websocket"
	"net/url"
)

type AgentConfig struct {
	AgentOptions *config.AgentOptions
	Container    *container.Container
	WebSocket    *websocket.WebSocket
	RequestChan  chan *utils.Request
	ResponseChan chan *utils.TResponse
}

func NewAgentConfig(opt *config.AgentOptions) (*AgentConfig, error) {
	agentConfig := &AgentConfig{
		AgentOptions: opt,
		RequestChan:  make(chan *utils.Request),
		ResponseChan: make(chan *utils.TResponse),
	}

	serverUrl := &url.URL{Scheme: "ws", Host: opt.ServerUrl, Path: "/osp/kube/connect"}
	agentConfig.WebSocket = websocket.NewWebSocket(
		serverUrl,
		opt.AgentToken,
		agentConfig.RequestChan,
		agentConfig.ResponseChan)

	kubeClient := kubernetes.NewKubeClient(opt.KubeConfigFile)
	agentConfig.Container = container.NewContainer(
		//nil,
		kubeClient,
		agentConfig.RequestChan,
		agentConfig.ResponseChan,
		agentConfig.WebSocket.SendResponse)

	return agentConfig, nil
}

type Agent struct {
	Container *container.Container
	WebSocket *websocket.WebSocket
}

func NewAgent(config *AgentConfig) *Agent {
	return &Agent{
		Container: config.Container,
		WebSocket: config.WebSocket,
	}
}

func (a *Agent) Run() {
	go a.WebSocket.ReadRequest()
	go a.WebSocket.WriteResponse()
	a.Container.Run()
}
