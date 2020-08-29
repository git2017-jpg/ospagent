package kubernetes

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type InformerRegistry interface {
	PodInformer() v1.PodInformer
	NamespaceInformer() v1.NamespaceInformer
	NodeInformer() v1.NodeInformer
	PersistentVolumeInformer() v1.PersistentVolumeInformer
	ConfigMapInformer() v1.ConfigMapInformer
}

type InformerRegistryImpl struct {
	podInformer              v1.PodInformer
	nameSpaceInformer        v1.NamespaceInformer
	nodeInformer             v1.NodeInformer
	persistentVolumeInformer v1.PersistentVolumeInformer
	configMapInformer        v1.ConfigMapInformer
}

func NewInformerRegistry(kubeClient kubernetes.Interface, stopCh <-chan struct{}) (InformerRegistry, error) {
	// 初始化 informer
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	defer runtime.HandleCrash()
	podInformer, err := NewPodInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}
	nsInformer, err := NewNamespaceInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}
	nodeInformer, err := NewNodeInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}
	persistentVolumeInformer, err := NewPersistentVolumeInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}
	configMapInformer, err := NewConfigMapInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}

	return &InformerRegistryImpl{
		podInformer:              podInformer,
		nameSpaceInformer:        nsInformer,
		nodeInformer:             nodeInformer,
		persistentVolumeInformer: persistentVolumeInformer,
		configMapInformer:        configMapInformer,
	}, nil
}

func NewPodInformer(factory informers.SharedInformerFactory, stopCh <-chan struct{}) (v1.PodInformer, error) {
	podInformer := factory.Core().V1().Pods()
	informer := podInformer.Informer()
	defer runtime.HandleCrash()

	// 启动 informer，list & watch
	factory.Start(stopCh)
	//从 apiserver 同步资源，即 list
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return nil, fmt.Errorf("timed out waiting for caches to sync")
	}
	return podInformer, nil
}

func NewNamespaceInformer(factory informers.SharedInformerFactory, stopCh <-chan struct{}) (v1.NamespaceInformer, error) {
	nsInformer := factory.Core().V1().Namespaces()
	informer := nsInformer.Informer()
	defer runtime.HandleCrash()

	// 启动 informer，list & watch
	factory.Start(stopCh)
	//从 apiserver 同步资源，即 list
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return nil, fmt.Errorf("timed out waiting for caches to sync")
	}
	return nsInformer, nil
}

func NewNodeInformer(factory informers.SharedInformerFactory, stopCh <-chan struct{}) (v1.NodeInformer, error) {
	nodeInformer := factory.Core().V1().Nodes()
	informer := nodeInformer.Informer()
	defer runtime.HandleCrash()

	factory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("time out waiting for caches to sync"))
		return nil, fmt.Errorf("time out waiting for caches to sync")
	}
	return nodeInformer, nil
}

func NewPersistentVolumeInformer(factory informers.SharedInformerFactory, stopCh <-chan struct{}) (v1.PersistentVolumeInformer, error) {
	persistentVolumeInformer := factory.Core().V1().PersistentVolumes()
	informer := persistentVolumeInformer.Informer()
	defer runtime.HandleCrash()

	factory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("time out waiting for caches to sync"))
		return nil, fmt.Errorf("time out waiting for caches to sync")
	}
	return persistentVolumeInformer, nil
}

func NewConfigMapInformer(factory informers.SharedInformerFactory, stopCh <-chan struct{}) (v1.ConfigMapInformer, error) {
	configMapInformer := factory.Core().V1().ConfigMaps()
	informer := configMapInformer.Informer()
	defer runtime.HandleCrash()

	factory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("time out waiting for caches to sync"))
		return nil, fmt.Errorf("time out waiting for caches to sync")
	}
	return configMapInformer, nil
}

func (r *InformerRegistryImpl) PodInformer() v1.PodInformer {
	return r.podInformer
}

func (r *InformerRegistryImpl) NamespaceInformer() v1.NamespaceInformer {
	return r.nameSpaceInformer
}

func (r *InformerRegistryImpl) NodeInformer() v1.NodeInformer {
	return r.nodeInformer
}

func (r *InformerRegistryImpl) PersistentVolumeInformer() v1.PersistentVolumeInformer {
	return r.persistentVolumeInformer
}

func (r *InformerRegistryImpl) ConfigMapInformer() v1.ConfigMapInformer {
	return r.configMapInformer
}
