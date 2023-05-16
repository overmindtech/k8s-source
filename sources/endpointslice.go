package sources

import (
	v1 "k8s.io/api/discovery/v1"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func endpointSliceExtractor(resource *v1.EndpointSlice, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	sd, err := ParseScope(scope, true)

	if err != nil {
		return nil, err
	}

	for _, endpoint := range resource.Endpoints {
		if endpoint.Hostname != nil {
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_GET,
					Query:  *endpoint.Hostname,
					Scope:  "global",
				},
			})
		}

		if endpoint.NodeName != nil {
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "Node",
					Method: sdp.QueryMethod_GET,
					Query:  *endpoint.NodeName,
					Scope:  sd.ClusterName,
				},
			})
		}

		if endpoint.TargetRef != nil {
			newQuery := ObjectReferenceToQuery(endpoint.TargetRef, sd)
			queries = append(queries, newQuery)
		}

		for _, address := range endpoint.Addresses {
			switch resource.AddressType {
			case v1.AddressTypeIPv4, v1.AddressTypeIPv6:
				queries = append(queries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  address,
						Scope:  "global",
					},
				})
			case v1.AddressTypeFQDN:
				queries = append(queries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "dns",
						Method: sdp.QueryMethod_GET,
						Query:  address,
						Scope:  "global",
					},
				})
			}
		}
	}

	return queries, nil
}

func newEndpointSliceSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.EndpointSlice, *v1.EndpointSliceList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "EndpointSlice",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.EndpointSlice, *v1.EndpointSliceList] {
			return cs.DiscoveryV1().EndpointSlices(namespace)
		},
		ListExtractor: func(list *v1.EndpointSliceList) ([]*v1.EndpointSlice, error) {
			extracted := make([]*v1.EndpointSlice, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: endpointSliceExtractor,
	}
}

func init() {
	registerSourceLoader(newEndpointSliceSource)
}
