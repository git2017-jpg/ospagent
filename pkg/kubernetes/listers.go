package kubernetes

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"time"
)

type ListRegistry interface {
	PodLister() v1.PodLister
}

type ListRegistryImpl struct {
	podLister v1.PodLister
}

func NewListRegistry(kubeClient kubernetes.Interface, stopCh <-chan struct{}) ListRegistry {
	podLister := NewPodLister(kubeClient, stopCh)
	return &ListRegistryImpl{podLister: podLister}
}

func (r *ListRegistryImpl) PodLister() v1.PodLister {
	return r.podLister
}

func NewPodLister(kubeClient kubernetes.Interface, stopCh <-chan struct{}) v1.PodLister {
	podListWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", "", fields.Everything())
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(podListWatch, &apiv1.Pod{}, time.Hour)
	podLister := v1.NewPodLister(store)
	go reflector.Run(stopCh)
	return podLister
}
