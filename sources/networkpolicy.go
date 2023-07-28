package sources

import (
	v1 "k8s.io/api/networking/v1"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func NetworkPolicyExtractor(resource *v1.NetworkPolicy, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	// +overmind:link Pod
	queries = append(queries, &sdp.LinkedItemQuery{
		Query: &sdp.Query{
			Type:   "Pod",
			Method: sdp.QueryMethod_SEARCH,
			Query:  LabelSelectorToQuery(&resource.Spec.PodSelector),
			Scope:  scope,
		},
		BlastPropagation: &sdp.BlastPropagation{
			// Changes to pods won't affect the network policy or anything else
			// that shares it
			In: false,
			// Changes to the network policy will affect pods
			Out: true,
		},
	})

	var peers []v1.NetworkPolicyPeer

	for _, ig := range resource.Spec.Ingress {
		peers = append(peers, ig.From...)
	}

	for _, eg := range resource.Spec.Egress {
		peers = append(peers, eg.To...)
	}

	// Link all peers
	for _, peer := range peers {
		if ps := peer.PodSelector; ps != nil {
			// TODO: Link to namespaces that are allowed to ingress e.g.
			// - namespaceSelector:
			// matchLabels:
			//   project: something

			// +overmind:link Pod
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Scope:  scope,
					Method: sdp.QueryMethod_GET,
					Query:  LabelSelectorToQuery(ps),
					Type:   "Pod",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to pods won't affect the network policy or anything else
					// that shares it
					In: false,
					// Changes to the network policy will affect pods
					Out: true,
				},
			})
		}
	}

	return queries, nil
}

//go:generate docgen ../docs-data
// +overmind:type NetworkPolicy
// +overmind:descriptiveType Network Policy
// +overmind:get Get a network policy by name
// +overmind:list List all network policies
// +overmind:search Search for a network policy using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_network_policy.metadata.name
// +overmind:terraform:queryMap kubernetes_network_policy_v1.metadata.name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}

func newNetworkPolicySource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.NetworkPolicy, *v1.NetworkPolicyList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "NetworkPolicy",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.NetworkPolicy, *v1.NetworkPolicyList] {
			return cs.NetworkingV1().NetworkPolicies(namespace)
		},
		ListExtractor: func(list *v1.NetworkPolicyList) ([]*v1.NetworkPolicy, error) {
			extracted := make([]*v1.NetworkPolicy, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: NetworkPolicyExtractor,
	}
}

func init() {
	registerSourceLoader(newNetworkPolicySource)
}
