package resource

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
	"github.com/openspacee/ospagent/pkg/utils/code"
	"github.com/openspacee/ospagent/pkg/websocket"
	"io"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog"
	"sigs.k8s.io/yaml"
	"strings"
)

type Pod struct {
	*kubernetes.KubeClient
	websocket.SendResponse
	watch        *WatchResource
	execSessions map[string]*streamHandler
	logSessions  map[string]*logHandler
}

func NewPod(kubeClient *kubernetes.KubeClient, sendResponse websocket.SendResponse, watch *WatchResource) *Pod {
	pod := &Pod{
		KubeClient:   kubeClient,
		SendResponse: sendResponse,
		watch:        watch,
		execSessions: make(map[string]*streamHandler),
		logSessions:  make(map[string]*logHandler),
	}
	pod.DoWatch()
	return pod
}

type PodQueryParams struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Output    string `json:"output"`
}

type WatchPodParams struct {
	UID string `json:"uid"`
}

type BuildContainer struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Restarts int32  `json:"restarts"`
	Ready    bool   `json:"ready"`
}

type BuildPod struct {
	UID             string            `json:"uid"`
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Containers      []*BuildContainer `json:"containers"`
	InitContainers  []*BuildContainer `json:"init_containers"`
	Controlled      string            `json:"controlled"`
	Qos             string            `json:"qos"`
	Created         string            `json:"created"`
	Status          string            `json:"status"`
	Ip              string            `json:"ip"`
	NodeName        string            `json:"node_name"`
	ResourceVersion string            `json:"resource_version"`
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
		UID:             string(pod.UID),
		Name:            pod.Name,
		Namespace:       pod.Namespace,
		Containers:      containers,
		InitContainers:  initContainers,
		Controlled:      controlled,
		Qos:             string(pod.Status.QOSClass),
		Status:          string(pod.Status.Phase),
		Ip:              pod.Status.PodIP,
		Created:         fmt.Sprint(pod.CreationTimestamp),
		NodeName:        pod.Spec.NodeName,
		ResourceVersion: pod.ResourceVersion,
	}
}

func (p *Pod) List(requestParams interface{}) *utils.Response {
	queryParams := &PodQueryParams{}
	json.Unmarshal(requestParams.([]byte), queryParams)
	podList, err := p.KubeClient.InformerRegistry.PodInformer().Lister().List(labels.Everything())
	if err != nil {
		return &utils.Response{
			Code: code.ListError,
			Msg:  err.Error(),
		}
	}
	var podRes []*BuildPod
	for _, pod := range podList {
		if queryParams.Name == "" || strings.Contains(pod.ObjectMeta.Name, queryParams.Name) {
			podRes = append(podRes, p.ToBuildPod(pod))
		}
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: podRes}
}

func (p *Pod) Get(requestParams interface{}) *utils.Response {
	queryParams := &PodQueryParams{}
	json.Unmarshal(requestParams.([]byte), queryParams)
	if queryParams.Name == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Pod name is blank"}
	}
	if queryParams.Namespace == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "Namespace is blank"}
	}
	pod, err := p.KubeClient.PodLister().Pods(queryParams.Namespace).Get(queryParams.Name)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	if queryParams.Output == "yaml" {
		podYaml, err := yaml.Marshal(pod)
		if err != nil {
			return &utils.Response{Code: code.MarshalError, Msg: err.Error()}
		}
		return &utils.Response{Code: code.Success, Msg: "Success", Data: string(podYaml)}
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: pod}
}

func (p *Pod) DoWatch() {
	podInformer := p.KubeClient.PodInformer().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    p.watch.WatchAdd(utils.WatchPod),
		UpdateFunc: p.watch.WatchUpdate(utils.WatchPod),
		DeleteFunc: p.watch.WatchDelete(utils.WatchPod),
	})
}

type PodExecParams struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Container string `json:"container"`
	SessionId string `json:"session_id"`
	Rows      string `json:"rows"`
	Cols      string `json:"cols"`
}

func (p *Pod) Exec(requestParams interface{}) *utils.Response {
	params := &PodExecParams{}
	json.Unmarshal(requestParams.([]byte), params)
	klog.Info(params)
	go p.startProcess(params.Name, params.Namespace, params.Container, params.SessionId, params.Rows, params.Cols)
	return &utils.Response{Code: code.Success, Msg: "Success"}
}

func (p *Pod) startProcess(podName, namespace, container, sessionId, rows, cols string) {
	execCmd := []string{"/bin/sh", "-c",
		fmt.Sprintf(`export LINES=%s; export COLUMNS=%s; 
	 TERM=xterm-256color; export TERM;
	 [ -x /bin/bash ] && ([ -x /usr/bin/script ] && /usr/bin/script -q -c \"/bin/bash\" /dev/null || exec /bin/bash) || exec /bin/sh`,
			rows, cols)}
	klog.Info(execCmd)
	sshReq := p.ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   execCmd,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(p.Config, "POST", sshReq.URL())
	if err != nil {
		klog.Error("exec pod container error", err)
		p.SendResponse(base64.StdEncoding.EncodeToString([]byte(err.Error())), sessionId, utils.ExecType)
		return
	}

	handler := &streamHandler{
		SessionId:    sessionId,
		resizeEvent:  make(chan remotecommand.TerminalSize),
		InChan:       make(chan []byte),
		SendResponse: p.SendResponse,
	}
	klog.Info("start stream session", sessionId)
	p.execSessions[sessionId] = handler
	defer func() {
		delete(p.execSessions, sessionId)
	}()
	if err := executor.Stream(remotecommand.StreamOptions{
		Stdin:             handler,
		Stdout:            handler,
		Stderr:            handler,
		TerminalSizeQueue: handler,
		Tty:               true,
	}); err != nil {
		klog.Errorf("exec pod container error session %s: %v", sessionId, err)
		p.SendResponse(base64.StdEncoding.EncodeToString([]byte(err.Error())), sessionId, utils.ExecType)
		return
	}
	klog.Info("end stream session", sessionId)
}

type StdInParams struct {
	SessionId string `json:"session_id"`
	Input     string `json:"input"`
	Width     uint16 `json:"width"`
	Height    uint16 `json:"height"`
}

func (p *Pod) ExecStdIn(requestParams interface{}) *utils.Response {
	params := &StdInParams{}
	json.Unmarshal(requestParams.([]byte), params)
	handler := p.execSessions[params.SessionId]
	if handler == nil {
		return &utils.Response{Code: code.ParamsError, Msg: "Not found session id"}
	}
	if params.Width > 0 && params.Height > 0 {
		handler.resizeEvent <- remotecommand.TerminalSize{Width: params.Width, Height: params.Height}
	}
	handler.InChan <- []byte(params.Input)
	return &utils.Response{Code: code.Success, Msg: "Success"}
}

type streamHandler struct {
	SessionId string
	InChan    chan []byte
	websocket.SendResponse
	resizeEvent chan remotecommand.TerminalSize
}

func (s *streamHandler) Read(p []byte) (size int, err error) {
	select {
	case inData, ok := <-s.InChan:
		if ok {
			size = len(inData)
			copy(p, inData)
		}
	}
	return
}

func (s *streamHandler) Write(p []byte) (size int, err error) {
	copyData := make([]byte, len(p))
	copy(copyData, p)
	size = len(p)
	s.SendResponse(copyData, s.SessionId, utils.ExecType)
	return
}

// executor回调获取web是否resize
func (s *streamHandler) Next() (size *remotecommand.TerminalSize) {
	ret := <-s.resizeEvent
	size = &ret
	return
}

type OpenPodLogParams struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Container string `json:"container"`
	SessionId string `json:"session_id"`
}

func (p *Pod) OpenLog(requestParams interface{}) *utils.Response {
	params := &OpenPodLogParams{}
	json.Unmarshal(requestParams.([]byte), params)
	klog.Info(params)
	tailLines := int64(100)
	podLogOpts := &v1.PodLogOptions{
		Container: params.Container,
		Follow:    true,
		TailLines: &tailLines,
		//Timestamps: true,
	}
	go p.logProcess(params.Namespace, params.Name, params.SessionId, podLogOpts)
	return &utils.Response{Code: code.Success, Msg: "Success"}
}

type ClosePodLogParams struct {
	SessionId string `json:"session_id"`
}

func (p *Pod) CloseLog(requestParams interface{}) *utils.Response {
	params := &ClosePodLogParams{}
	json.Unmarshal(requestParams.([]byte), params)
	klog.Info(params)
	handler := p.logSessions[params.SessionId]
	if handler != nil {
		klog.Info("close log session ", handler.SessionId)
		handler.PodLogs.Close()
	}
	return &utils.Response{Code: code.Success, Msg: "Success"}
}

type logHandler struct {
	SessionId string
	websocket.SendResponse
	PodLogs io.ReadCloser
}

func (l *logHandler) Write(p []byte) (size int, err error) {
	copyData := make([]byte, len(p))
	copy(copyData, p)
	size = len(p)
	l.SendResponse(copyData, l.SessionId, utils.LogType)
	return
}

func (p *Pod) logProcess(namespace, name, sessionId string, podLogOpts *v1.PodLogOptions) {

	req := p.ClientSet.CoreV1().Pods(namespace).GetLogs(name, podLogOpts)
	podLogs, err := req.Stream()
	if err != nil {
		klog.Errorf("open log stream session %s error: %v", sessionId, err)
		p.SendResponse(base64.StdEncoding.EncodeToString([]byte(err.Error())), sessionId, utils.LogType)
		return
	}
	defer podLogs.Close()

	handler := &logHandler{
		SessionId:    sessionId,
		SendResponse: p.SendResponse,
		PodLogs:      podLogs,
	}
	klog.Info("start log session ", sessionId)
	p.logSessions[sessionId] = handler
	defer func() {
		delete(p.logSessions, sessionId)
	}()

	_, err = io.Copy(handler, podLogs)
	if err != nil {
		klog.Errorf("copy log session %s error: %v", sessionId, err)
		p.SendResponse(base64.StdEncoding.EncodeToString([]byte(err.Error())), sessionId, utils.LogType)
		return
	}
	klog.Info("end log session ", sessionId)
}
