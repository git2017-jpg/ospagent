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
	ConfigMapLister() v1.ConfigMapLister
}

type ListRegistryImpl struct {
	podLister       v1.PodLister
	configMapLister v1.ConfigMapLister
}

func NewListRegistry(kubeClient kubernetes.Interface, stopCh <-chan struct{}) ListRegistry {
	podLister := NewPodLister(kubeClient, stopCh)
	configMapLister := NewConfigMapLister(kubeClient, stopCh)

	return &ListRegistryImpl{
		podLister:       podLister,
		configMapLister: configMapLister,
	}
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

func (r *ListRegistryImpl) ConfigMapLister() v1.ConfigMapLister {
	return r.configMapLister
}

func NewConfigMapLister(kubeClient kubernetes.Interface, stopCh <-chan struct{}) v1.ConfigMapLister {
	configMapListWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "configMaps", "", fields.Everything())
	store, reflector := cache.NewNamespaceKeyedIndexerAndReflector(configMapListWatch, &apiv1.ConfigMap{}, time.Hour)
	configMapLister := v1.NewConfigMapLister(store)
	go reflector.Run(stopCh)
	return configMapLister
}
