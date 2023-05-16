package sources

import (
	"github.com/overmindtech/discovery"
	v1 "k8s.io/api/apps/v1"

	"k8s.io/client-go/kubernetes"
)

func newDaemonSetSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.DaemonSet, *v1.DaemonSetList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "DaemonSet",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.DaemonSet, *v1.DaemonSetList] {
			return cs.AppsV1().DaemonSets(namespace)
		},
		ListExtractor: func(list *v1.DaemonSetList) ([]*v1.DaemonSet, error) {
			extracted := make([]*v1.DaemonSet, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		// Pods are linked automatically
	}
}

func init() {
	registerSourceLoader(newDaemonSetSource)
}
