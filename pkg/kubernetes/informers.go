package kubernetes

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	appsv1 "k8s.io/client-go/informers/apps/v1"
	hpa "k8s.io/client-go/informers/autoscaling/v1"
	batchv1 "k8s.io/client-go/informers/batch/v1"
	batchv1beta1 "k8s.io/client-go/informers/batch/v1beta1"
	"k8s.io/client-go/informers/core/v1"
	storage "k8s.io/client-go/informers/storage/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type InformerRegistry interface {
	PodInformer() v1.PodInformer
	NamespaceInformer() v1.NamespaceInformer
	NodeInformer() v1.NodeInformer
	EventInformer() v1.EventInformer
	DeploymentInformer() appsv1.DeploymentInformer
	PersistentVolumeInformer() v1.PersistentVolumeInformer
	PersistentVolumeClaimInformer() v1.PersistentVolumeClaimInformer
	StorageClassInformer() storage.StorageClassInformer
	ConfigMapInformer() v1.ConfigMapInformer
	StatefulSetInformer() appsv1.StatefulSetInformer
	DaemonSetInformer() appsv1.DaemonSetInformer
	JobInformer() batchv1.JobInformer
	CronJobInformer() batchv1beta1.CronJobInformer
	HorizontalPodAutoscalerInformer() hpa.HorizontalPodAutoscalerInformer
}

type InformerRegistryImpl struct {
	podInformer                     v1.PodInformer
	nameSpaceInformer               v1.NamespaceInformer
	nodeInformer                    v1.NodeInformer
	eventInformer                   v1.EventInformer
	deploymentInformer              appsv1.DeploymentInformer
	persistentVolumeInformer        v1.PersistentVolumeInformer
	persistentVolumeClaimInformer   v1.PersistentVolumeClaimInformer
	storageClassInformer            storage.StorageClassInformer
	configMapInformer               v1.ConfigMapInformer
	statefulSetInformer             appsv1.StatefulSetInformer
	daemonSetInformer               appsv1.DaemonSetInformer
	jobInformer                     batchv1.JobInformer
	cronJobInformer                 batchv1beta1.CronJobInformer
	horizontalPodAutoscalerInformer hpa.HorizontalPodAutoscalerInformer
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
	persistentVolumeClaimInformer, err := NewPersistentVolumeClaimInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}
	configMapInformer, err := NewConfigMapInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}

	eventInformer, err := NewEventInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}

	deploymentInformer, err := NewDeploymentInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}

	statefulSetInformer, err := NewStatefulSetInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}

	daemonSetInformer, err := NewDaemonSetInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}

	jobInformer, err := NewJobInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}

	cronJobInformer, err := NewCronJobInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}

	storageClassInformer, err := NewStorageClassInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}

	horizontalPodAutoscalerInformer, err := NewHorizontalPodAutoscalerInformer(factory, stopCh)
	if err != nil {
		return nil, err
	}

	return &InformerRegistryImpl{
		podInformer:                     podInformer,
		nameSpaceInformer:               nsInformer,
		nodeInformer:                    nodeInformer,
		eventInformer:                   eventInformer,
		deploymentInformer:              deploymentInformer,
		persistentVolumeInformer:        persistentVolumeInformer,
		persistentVolumeClaimInformer:   persistentVolumeClaimInformer,
		configMapInformer:               configMapInformer,
		statefulSetInformer:             statefulSetInformer,
		daemonSetInformer:               daemonSetInformer,
		jobInformer:                     jobInformer,
		cronJobInformer:                 cronJobInformer,
		storageClassInformer:            storageClassInformer,
		horizontalPodAutoscalerInformer: horizontalPodAutoscalerInformer,
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

func NewEventInformer(factory informers.SharedInformerFactory, stopCh <-chan struct{}) (v1.EventInformer, error) {
	eventInformer := factory.Core().V1().Events()
	informer := eventInformer.Informer()
	defer runtime.HandleCrash()

	factory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("time out waiting for caches to sync"))
		return nil, fmt.Errorf("time out waiting for caches to sync")
	}
	return eventInformer, nil
}

func NewDeploymentInformer(factory informers.SharedInformerFactory, stopCh <-chan struct{}) (appsv1.DeploymentInformer, error) {
	deploymentInformer := factory.Apps().V1().Deployments()
	informer := deploymentInformer.Informer()
	defer runtime.HandleCrash()

	factory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("time out waiting for caches to sync"))
		return nil, fmt.Errorf("time out waiting for caches to sync")
	}
	return deploymentInformer, nil
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

func NewPersistentVolumeClaimInformer(factory informers.SharedInformerFactory, stopCh <-chan struct{}) (v1.PersistentVolumeClaimInformer, error) {
	persistentVolumeClaimInformer := factory.Core().V1().PersistentVolumeClaims()
	informer := persistentVolumeClaimInformer.Informer()
	defer runtime.HandleCrash()

	factory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("time out waiting for caches to sync"))
		return nil, fmt.Errorf("time out waiting for caches to sync")
	}
	return persistentVolumeClaimInformer, nil
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

func NewStatefulSetInformer(factory informers.SharedInformerFactory, stopCh <-chan struct{}) (appsv1.StatefulSetInformer, error) {
	statefulSetInformer := factory.Apps().V1().StatefulSets()
	informer := statefulSetInformer.Informer()
	defer runtime.HandleCrash()

	factory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("time out waiting for caches to sync"))
		return nil, fmt.Errorf("time out waiting for caches to sync")
	}
	return statefulSetInformer, nil
}

func NewDaemonSetInformer(factory informers.SharedInformerFactory, stopCh <-chan struct{}) (appsv1.DaemonSetInformer, error) {
	daemonSetInformer := factory.Apps().V1().DaemonSets()
	informer := daemonSetInformer.Informer()
	defer runtime.HandleCrash()

	factory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("time out waiting for caches to sync"))
		return nil, fmt.Errorf("time out waiting for caches to sync")
	}
	return daemonSetInformer, nil
}

func NewJobInformer(factory informers.SharedInformerFactory, stopCh <-chan struct{}) (batchv1.JobInformer, error) {
	jobInformer := factory.Batch().V1().Jobs()
	informer := jobInformer.Informer()
	defer runtime.HandleCrash()

	factory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("time out waiting for caches to sync"))
		return nil, fmt.Errorf("time out waiting for caches to sync")
	}
	return jobInformer, nil
}

func NewCronJobInformer(factory informers.SharedInformerFactory, stopCh <-chan struct{}) (batchv1beta1.CronJobInformer, error) {
	cronJobInformer := factory.Batch().V1beta1().CronJobs()
	informer := cronJobInformer.Informer()
	defer runtime.HandleCrash()

	factory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("time out waiting for caches to sync"))
		return nil, fmt.Errorf("time out waiting for caches to sync")
	}
	return cronJobInformer, nil
}

func NewStorageClassInformer(factory informers.SharedInformerFactory, stopCh <-chan struct{}) (storage.StorageClassInformer, error) {
	storageClassInformer := factory.Storage().V1().StorageClasses()
	informer := storageClassInformer.Informer()
	defer runtime.HandleCrash()

	factory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("time out waiting for caches to sync"))
		return nil, fmt.Errorf("time out waiting for caches to sync")
	}
	return storageClassInformer, nil
}

func NewHorizontalPodAutoscalerInformer(factory informers.SharedInformerFactory, stopCh <-chan struct{}) (hpa.HorizontalPodAutoscalerInformer, error) {
	horizontalPodAutoscalerInformer := factory.Autoscaling().V1().HorizontalPodAutoscalers()
	informer := horizontalPodAutoscalerInformer.Informer()
	defer runtime.HandleCrash()

	factory.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("time out waiting for caches to sync"))
		return nil, fmt.Errorf("time out waiting for caches to sync")
	}
	return horizontalPodAutoscalerInformer, nil
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

func (r *InformerRegistryImpl) EventInformer() v1.EventInformer {
	return r.eventInformer
}

func (r *InformerRegistryImpl) DeploymentInformer() appsv1.DeploymentInformer {
	return r.deploymentInformer
}

func (r *InformerRegistryImpl) PersistentVolumeInformer() v1.PersistentVolumeInformer {
	return r.persistentVolumeInformer
}

func (r *InformerRegistryImpl) PersistentVolumeClaimInformer() v1.PersistentVolumeClaimInformer {
	return r.persistentVolumeClaimInformer
}

func (r *InformerRegistryImpl) ConfigMapInformer() v1.ConfigMapInformer {
	return r.configMapInformer
}

func (r *InformerRegistryImpl) StatefulSetInformer() appsv1.StatefulSetInformer {
	return r.statefulSetInformer
}

func (r *InformerRegistryImpl) DaemonSetInformer() appsv1.DaemonSetInformer {
	return r.daemonSetInformer
}

func (r *InformerRegistryImpl) JobInformer() batchv1.JobInformer {
	return r.jobInformer
}

func (r *InformerRegistryImpl) CronJobInformer() batchv1beta1.CronJobInformer {
	return r.cronJobInformer
}

func (r *InformerRegistryImpl) StorageClassInformer() storage.StorageClassInformer {
	return r.storageClassInformer
}

func (r *InformerRegistryImpl) HorizontalPodAutoscalerInformer() hpa.HorizontalPodAutoscalerInformer {
	return r.horizontalPodAutoscalerInformer
}
