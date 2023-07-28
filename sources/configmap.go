package sources

import (
	"github.com/overmindtech/discovery"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

//go:generate docgen ../docs-data
// +overmind:type ConfigMap
// +overmind:descriptiveType Config Map
// +overmind:get Get a config map by name
// +overmind:list List all config maps
// +overmind:search Search for a config map using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_config_map.metadata.name
// +overmind:terraform:queryMap kubernetes_config_map_v1.metadata.name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}

func newConfigMapSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.ConfigMap, *v1.ConfigMapList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "ConfigMap",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.ConfigMap, *v1.ConfigMapList] {
			return cs.CoreV1().ConfigMaps(namespace)
		},
		ListExtractor: func(list *v1.ConfigMapList) ([]*v1.ConfigMap, error) {
			bindings := make([]*v1.ConfigMap, len(list.Items))

			for i := range list.Items {
				bindings[i] = &list.Items[i]
			}

			return bindings, nil
		},
	}
}

func init() {
	registerSourceLoader(newConfigMapSource)
}
