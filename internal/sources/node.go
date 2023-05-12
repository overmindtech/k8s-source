package sources

import (
	v1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
)

func NewNodeSource(cs *kubernetes.Clientset, cluster string, namespaces []string) KubeTypeSource[*v1.Node, *v1.NodeList] {
	return KubeTypeSource[*v1.Node, *v1.NodeList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "Node",
		ClusterInterfaceBuilder: func() ItemInterface[*v1.Node, *v1.NodeList] {
			return cs.CoreV1().Nodes()
		},
		ListExtractor: func(list *v1.NodeList) ([]*v1.Node, error) {
			extracted := make([]*v1.Node, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
	}
}
