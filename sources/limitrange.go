package sources

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func newLimitRangeSource(cs *kubernetes.Clientset, cluster string, namespaces []string) *KubeTypeSource[*v1.LimitRange, *v1.LimitRangeList] {
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
