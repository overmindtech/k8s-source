package sources

import (
	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func NewConfigMapSource(cs *kubernetes.Clientset, cluster string, namespaces []string) KubeTypeSource[*v1.ConfigMap, *v1.ConfigMapList] {
	return KubeTypeSource[*v1.ConfigMap, *v1.ConfigMapList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "ConfigMap",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.ConfigMap, *v1.ConfigMapList] {
			return cs.CoreV1().ConfigMaps(namespace)
		},
		ListExtractor: func(list *v1.ConfigMapList) ([]*v1.ConfigMap, error) {
			bindings := make([]*v1.ConfigMap, len(list.Items))

			for i, crb := range list.Items {
				bindings[i] = &crb
			}

			return bindings, nil
		},
		LinkedItemQueryExtractor: func(resource *v1.ConfigMap, scope string) ([]*sdp.Query, error) {
			return []*sdp.Query{}, nil
		},
	}
}
