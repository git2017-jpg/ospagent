package resource

import (
	"encoding/json"
	"fmt"
	"github.com/openspacee/ospagent/pkg/kubernetes"
	"github.com/openspacee/ospagent/pkg/utils"
	"github.com/openspacee/ospagent/pkg/utils/code"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

type Event struct {
	*kubernetes.KubeClient
	watch *WatchResource
}

func NewEvent(kubeClient *kubernetes.KubeClient, watch *WatchResource) *Event {
	e := &Event{
		KubeClient: kubeClient,
		watch:      watch,
	}
	e.DoWatch()
	return e
}

func (e *Event) DoWatch() {
	eventInformer := e.KubeClient.EventInformer().Informer()
	eventInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    e.watch.WatchAdd(utils.WatchEvent),
		UpdateFunc: e.watch.WatchUpdate(utils.WatchEvent),
		DeleteFunc: e.watch.WatchDelete(utils.WatchEvent),
	})
}

type BuildEvent struct {
	UID             string              `json:"uid"`
	Reason          string              `json:"reason"`
	Message         string              `json:"message"`
	Type            string              `json:"type"`
	Object          *v1.ObjectReference `json:"object"`
	Source          *v1.EventSource     `json:"source"`
	EventTime       string              `json:"event_time"`
	ResourceVersion string              `json:"resource_version"`
}

func (e *Event) ToBuildEvent(event *v1.Event) *BuildEvent {
	if e == nil {
		return nil
	}
	eventTime := event.LastTimestamp
	if eventTime.IsZero() {
		eventTime = event.FirstTimestamp
	}
	if eventTime.IsZero() {
		eventTime = event.CreationTimestamp
	}
	eventData := &BuildEvent{
		UID:             string(event.UID),
		Reason:          event.Reason,
		Message:         event.Message,
		Type:            event.Type,
		Object:          &event.InvolvedObject,
		Source:          &event.Source,
		EventTime:       fmt.Sprint(eventTime.Format("2006-01-02T15:04:05Z")),
		ResourceVersion: event.ResourceVersion,
	}

	return eventData
}

type EventQueryParams struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	UID       string `json:"uid"`
}

func (e *Event) List(requestParams interface{}) *utils.Response {
	queryParams := &EventQueryParams{}
	json.Unmarshal(requestParams.([]byte), queryParams)
	eventList, err := e.KubeClient.InformerRegistry.EventInformer().Lister().List(labels.Everything())
	if err != nil {
		return &utils.Response{
			Code: code.ListError,
			Msg:  err.Error(),
		}
	}
	var events []*BuildEvent
	for _, event := range eventList {
		if queryParams.UID != "" && string(event.InvolvedObject.UID) != queryParams.UID {
			continue
		}
		if queryParams.Kind != "" && event.InvolvedObject.Kind != queryParams.Kind {
			continue
		}
		if queryParams.Namespace != "" && event.InvolvedObject.Namespace != queryParams.Namespace {
			continue
		}
		if queryParams.Name != "" && event.InvolvedObject.Name != queryParams.Name {
			continue
		}
		events = append(events, e.ToBuildEvent(event))
	}
	return &utils.Response{Code: code.Success, Msg: "Success", Data: events}
}
