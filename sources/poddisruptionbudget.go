package sources

import (
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/policy/v1"
	"k8s.io/client-go/kubernetes"
)

func podDisruptionBudgetExtractor(resource *v1.PodDisruptionBudget, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	if resource.Spec.Selector != nil {
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "Pod",
				Method: sdp.QueryMethod_SEARCH,
				Query:  LabelSelectorToQuery(resource.Spec.Selector),
				Scope:  scope,
			},
		})
	}

	return queries, nil
}

func newPodDisruptionBudgetSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.PodDisruptionBudget, *v1.PodDisruptionBudgetList]{
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

func init() {
	registerSourceLoader(newPodDisruptionBudgetSource)
}
