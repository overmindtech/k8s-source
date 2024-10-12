package adapters

import (
	"github.com/overmindtech/discovery"
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
				// +overmind:link DNS
				queries = append(queries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Scope:  "global",
						Method: sdp.QueryMethod_GET,
						Query:  address.Hostname,
						Type:   "dns",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Always propagate over DNS
						In:  true,
						Out: true,
					},
				})
			}

			if address.NodeName != nil {
				// +overmind:link Node
				queries = append(queries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "Node",
						Scope:  sd.ClusterName,
						Method: sdp.QueryMethod_GET,
						Query:  *address.NodeName,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to the node can affect the endpoint
						In: true,
						// Changes to the endpoint cannot affect the node
						Out: false,
					},
				})
			}

			if address.IP != "" {
				// +overmind:link ip
				queries = append(queries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ip",
						Method: sdp.QueryMethod_GET,
						Query:  address.IP,
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Always propagate over IP
						In:  true,
						Out: true,
					},
				})
			}

			if address.TargetRef != nil {
				// +overmind:link Pod
				// +overmind:link ExternalName
				queries = append(queries, ObjectReferenceToQuery(address.TargetRef, sd, &sdp.BlastPropagation{
					// These are tightly coupled
					In:  true,
					Out: true,
				}))
			}
		}
	}

	return queries, nil
}

//go:generate docgen ../docs-data
// +overmind:type Endpoints
// +overmind:descriptiveType Endpoints
// +overmind:get Get an endpoint by name
// +overmind:list List all endpoints
// +overmind:search Search for an endpoint using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_endpoints.metadata[0].name
// +overmind:terraform:queryMap kubernetes_endpoints_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}

func newEndpointsAdapter(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Adapter {
	return &KubeTypeAdapter[*v1.Endpoints, *v1.EndpointsList]{
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
		AdapterMetadata:          endpointsAdapterMetadata,
	}
}

var endpointsAdapterMetadata = AdapterMetadata.Register(&sdp.AdapterMetadata{
	DescriptiveName:       "Endpoints",
	Type:                  "Endpoints",
	Category:              sdp.AdapterCategory_ADAPTER_CATEGORY_NETWORK,
	SupportedQueryMethods: DefaultSupportedQueryMethods("Endpoints"),
	PotentialLinks:        []string{"Node", "ip", "Pod", "ExternalName", "DNS"},
	TerraformMappings: []*sdp.TerraformMapping{
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_endpoints.metadata[0].name",
		},
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_endpoints_v1.metadata[0].name",
		},
	},
})

func init() {
	registerAdapterLoader(newEndpointsAdapter)
}
