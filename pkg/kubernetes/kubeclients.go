package kubernetes

import (
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

type KubeClient struct {
	ClientSet kube_client.Interface
	ListRegistry
	InformerRegistry
}

func getKubeConfig(kubeConfigFile string) *rest.Config {
	if kubeConfigFile != "" {
		klog.Infof("using kubeconfig file: %s", kubeConfigFile)
		// use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", kubeConfigFile)
		if err != nil {
			panic("failed to build config: " + err.Error())
		}
		return config
	}

	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	return kubeConfig
}

func createKubeClient(kubeConfigFile string) kube_client.Interface {
	kubeConfig := getKubeConfig(kubeConfigFile)
	return kube_client.NewForConfigOrDie(kubeConfig)
}

func NewKubeClient(kubeConfigFile string) *KubeClient {
	kubeClient := createKubeClient(kubeConfigFile)
	listRegistry := NewListRegistry(kubeClient, nil)
	informerRegistry, err := NewInformerRegistry(kubeClient, nil)
	if err != nil {
		panic(err.Error())
	}
	return &KubeClient{
		ClientSet:        kubeClient,
		ListRegistry:     listRegistry,
		InformerRegistry: informerRegistry,
	}
}
