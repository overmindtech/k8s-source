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
			// +overmind:link dns
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_GET,
					Query:  *endpoint.Hostname,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Always propagate over DNS
					In:  true,
					Out: true,
				},
			})
		}

		if endpoint.NodeName != nil {
			// +overmind:link Node
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "Node",
					Method: sdp.QueryMethod_GET,
					Query:  *endpoint.NodeName,
					Scope:  sd.ClusterName,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the node can affect the endpoint
					In: true,
					// Changes to the endpoint cannot affect the node
					Out: false,
				},
			})
		}

		if endpoint.TargetRef != nil {
			// +overmind:link Pod
			queries = append(queries, ObjectReferenceToQuery(endpoint.TargetRef, sd, &sdp.BlastPropagation{
				// Changes to the pod could affect the endpoint and vice versa
				In:  true,
				Out: true,
			}))
		}

		for _, address := range endpoint.Addresses {
			switch resource.AddressType {
			case v1.AddressTypeIPv4, v1.AddressTypeIPv6:
				// +overmind:link ip
				queries = append(queries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  address,
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Always propagate over IP
						In:  true,
						Out: true,
					},
				})
			case v1.AddressTypeFQDN:
				// +overmind:link dns
				queries = append(queries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "dns",
						Method: sdp.QueryMethod_GET,
						Query:  address,
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Always propagate over DNS
						In:  true,
						Out: true,
					},
				})
			}
		}
	}

	return queries, nil
}

//go:generate docgen ../docs-data
// +overmind:type EndpointSlice
// +overmind:descriptiveType Endpoint Slice
// +overmind:get Get a endpoint slice by name
// +overmind:list List all endpoint slices
// +overmind:search Search for a endpoint slice using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_endpoints_slice_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}

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
