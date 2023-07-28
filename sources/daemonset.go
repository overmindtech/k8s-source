package sources

import (
	"github.com/overmindtech/discovery"
	v1 "k8s.io/api/apps/v1"

	"k8s.io/client-go/kubernetes"
)

//go:generate docgen ../docs-data
// +overmind:type DaemonSet
// +overmind:descriptiveType Daemon Set
// +overmind:get Get a daemon set by name
// +overmind:list List all daemon sets
// +overmind:search Search for a daemon set using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_daemonset.metadata[0].name
// +overmind:terraform:queryMap kubernetes_daemon_set_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}
// +overmind:link Pod

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
