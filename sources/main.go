package sources

import (
	"github.com/overmindtech/discovery"
	"k8s.io/client-go/kubernetes"
)

type SourceLoader func(clientSet *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source

var sourceLoaders []SourceLoader

func registerSourceLoader(loader SourceLoader) {
	sourceLoaders = append(sourceLoaders, loader)
}

func LoadAllSources(cs *kubernetes.Clientset, cluster string, namespaces []string) []discovery.Source {
	sources := make([]discovery.Source, len(sourceLoaders))

	for i, loader := range sourceLoaders {
		sources[i] = loader(cs, cluster, namespaces)
	}

	return sources
}
