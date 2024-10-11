package adapters

import (
	"github.com/overmindtech/discovery"
	"k8s.io/client-go/kubernetes"
)

type AdapterLoader func(clientSet *kubernetes.Clientset, cluster string, namespaces []string) discovery.Adapter

var adapterLoaders []AdapterLoader

func registerAdapterLoader(loader AdapterLoader) {
	adapterLoaders = append(adapterLoaders, loader)
}

func LoadAllAdapters(cs *kubernetes.Clientset, cluster string, namespaces []string) []discovery.Adapter {
	adapters := make([]discovery.Adapter, len(adapterLoaders))

	for i, loader := range adapterLoaders {
		adapters[i] = loader(cs, cluster, namespaces)
	}

	return adapters
}
