package sources

import (
	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/policy/v1"
	"k8s.io/client-go/kubernetes"
)

func podDisruptionBudgetExtractor(resource *v1.PodDisruptionBudget, scope string) ([]*sdp.Query, error) {
	queries := make([]*sdp.Query, 0)

	if resource.Spec.Selector != nil {
		queries = append(queries, &sdp.Query{
			Type:   "Pod",
			Method: sdp.QueryMethod_SEARCH,
			Query:  LabelSelectorToQuery(resource.Spec.Selector),
			Scope:  scope,
		})
	}

	return queries, nil
}

func NewPodDisruptionBudgetSource(cs *kubernetes.Clientset, cluster string, namespaces []string) KubeTypeSource[*v1.PodDisruptionBudget, *v1.PodDisruptionBudgetList] {
	return KubeTypeSource[*v1.PodDisruptionBudget, *v1.PodDisruptionBudgetList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "PodDisruptionBudget",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.PodDisruptionBudget, *v1.PodDisruptionBudgetList] {
			return cs.PolicyV1().PodDisruptionBudgets(namespace)
		},
		ListExtractor: func(list *v1.PodDisruptionBudgetList) ([]*v1.PodDisruptionBudget, error) {
			extracted := make([]*v1.PodDisruptionBudget, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: podDisruptionBudgetExtractor,
	}
}
