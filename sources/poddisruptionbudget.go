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
		// +overmind:link Pod
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "Pod",
				Method: sdp.QueryMethod_SEARCH,
				Query:  LabelSelectorToQuery(resource.Spec.Selector),
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changes to pods won't affect the disruption budget
				In: false,
				// Changes to the disruption budget will affect pods
				Out: true,
			},
		})
	}

	return queries, nil
}

//go:generate docgen ../docs-data
// +overmind:type PodDisruptionBudget
// +overmind:descriptiveType Pod Disruption Budget
// +overmind:get Get a pod disruption budget by name
// +overmind:list List all pod disruption budgets
// +overmind:search Search for a pod disruption budget using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_pod_disruption_budget_v1.metadata.name
// +overmind:terraform:scope ${outputs.overmind_kubernetes_cluster_name}.${values.metadata.namespace}

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
