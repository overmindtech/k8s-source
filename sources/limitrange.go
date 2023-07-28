package sources

import (
	"github.com/overmindtech/discovery"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

//go:generate docgen ../docs-data
// +overmind:type LimitRange
// +overmind:descriptiveType Limit Range
// +overmind:get Get a limit range by name
// +overmind:list List all limit ranges
// +overmind:search Search for a limit range using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_limit_range.metadata.name
// +overmind:terraform:queryMap kubernetes_limit_range_v1.metadata.name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata.namespace}

func newLimitRangeSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.LimitRange, *v1.LimitRangeList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "LimitRange",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.LimitRange, *v1.LimitRangeList] {
			return cs.CoreV1().LimitRanges(namespace)
		},
		ListExtractor: func(list *v1.LimitRangeList) ([]*v1.LimitRange, error) {
			extracted := make([]*v1.LimitRange, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
	}
}

func init() {
	registerSourceLoader(newLimitRangeSource)
}
