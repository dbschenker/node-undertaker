package nodeupdatehandler

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

func OnNodeUpdate(node *v1.Node) {

}

func GetDefaultUpdateHandlerFuncs() cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			OnNodeUpdate(newObj.(*v1.Node))
		},
		AddFunc: func(obj interface{}) {
			OnNodeUpdate(obj.(*v1.Node))
		},
		DeleteFunc: nil,
	}
}
