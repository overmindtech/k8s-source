package sources

import (
	"github.com/overmindtech/discovery"
	v1 "k8s.io/api/scheduling/v1"

	"k8s.io/client-go/kubernetes"
)

func newPriorityClassSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.PriorityClass, *v1.PriorityClassList]{
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

func init() {
	registerSourceLoader(newPriorityClassSource)
}
