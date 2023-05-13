package sources

import (
	v1 "k8s.io/api/scheduling/v1"

	"k8s.io/client-go/kubernetes"
)

func NewPriorityClassSource(cs *kubernetes.Clientset, cluster string, namespaces []string) KubeTypeSource[*v1.PriorityClass, *v1.PriorityClassList] {
	return KubeTypeSource[*v1.PriorityClass, *v1.PriorityClassList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "PriorityClass",
		ClusterInterfaceBuilder: func() ItemInterface[*v1.PriorityClass, *v1.PriorityClassList] {
			return cs.SchedulingV1().PriorityClasses()
		},
		ListExtractor: func(list *v1.PriorityClassList) ([]*v1.PriorityClass, error) {
			extracted := make([]*v1.PriorityClass, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
	}
}
