package sources

import (
	v1 "k8s.io/api/networking/v1"

	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func NetworkPolicyExtractor(resource *v1.NetworkPolicy, scope string) ([]*sdp.Query, error) {
	queries := make([]*sdp.Query, 0)

	queries = append(queries, &sdp.Query{
		Type:   "Pod",
		Method: sdp.QueryMethod_SEARCH,
		Query:  LabelSelectorToQuery(&resource.Spec.PodSelector),
		Scope:  scope,
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

			queries = append(queries, &sdp.Query{
				Scope:  scope,
				Method: sdp.QueryMethod_GET,
				Query:  LabelSelectorToQuery(ps),
				Type:   "Pod",
			})
		}
	}

	return queries, nil
}

func NewNetworkPolicySource(cs *kubernetes.Clientset, cluster string, namespaces []string) KubeTypeSource[*v1.NetworkPolicy, *v1.NetworkPolicyList] {
	return KubeTypeSource[*v1.NetworkPolicy, *v1.NetworkPolicyList]{
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
