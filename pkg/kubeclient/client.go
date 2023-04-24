package kubeclient

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func GetClient() (*kubernetes.Clientset, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		// Do something
	}
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}
