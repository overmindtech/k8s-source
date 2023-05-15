package sources

import (
	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func EndpointsExtractor(resource *v1.Endpoints, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	sd, err := ParseScope(scope, true)

	if err != nil {
		return nil, err
	}

	for _, subset := range resource.Subsets {
		for _, address := range subset.Addresses {
			if address.Hostname != "" {
				queries = append(queries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Scope:  "global",
						Method: sdp.QueryMethod_GET,
						Query:  address.Hostname,
						Type:   "dns",
					},
				})
			}

			if address.NodeName != nil {
				queries = append(queries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "Node",
						Scope:  sd.ClusterName,
						Method: sdp.QueryMethod_GET,
						Query:  *address.NodeName,
					},
				})
			}

			if address.IP != "" {
				queries = append(queries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  address.IP,
						Scope:  "global",
					},
				})
			}

			if address.TargetRef != nil {
				targetQuery := ObjectReferenceToQuery(address.TargetRef, sd)
				queries = append(queries, targetQuery)
			}
		}
	}

	return queries, nil
}

func newEndpointsSource(cs *kubernetes.Clientset, cluster string, namespaces []string) *KubeTypeSource[*v1.Endpoints, *v1.EndpointsList] {
	return &KubeTypeSource[*v1.Endpoints, *v1.EndpointsList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "Endpoints",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.Endpoints, *v1.EndpointsList] {
			return cs.CoreV1().Endpoints(namespace)
		},
		ListExtractor: func(list *v1.EndpointsList) ([]*v1.Endpoints, error) {
			extracted := make([]*v1.Endpoints, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: EndpointsExtractor,
	}
}
