package kubeclient

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClient - gets kubernetes client with namespace it runs in
func GetClient() (kubernetes.Interface, string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, "", err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, "", err
	}
	namespace, _, err := kubeConfig.Namespace()

	return clientset, namespace, err
}

func GetFakeClient() (kubernetes.Interface, string, error) {
	return fake.NewSimpleClientset(), metav1.NamespaceDefault, nil
}
