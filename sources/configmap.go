package sources

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func newConfigMapSource(cs *kubernetes.Clientset, cluster string, namespaces []string) *KubeTypeSource[*v1.ConfigMap, *v1.ConfigMapList] {
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
