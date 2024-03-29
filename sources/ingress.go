package sources

import (
	v1 "k8s.io/api/networking/v1"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func ingressExtractor(resource *v1.Ingress, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	if resource.Spec.IngressClassName != nil {
		// +overmind:link IngressClass
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "IngressClass",
				Method: sdp.QueryMethod_GET,
				Query:  *resource.Spec.IngressClassName,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changes to the ingress (e.g. nginx) class can affect the
				// ingresses that use it
				In: true,
				// Changes to an ingress wont' affect the class
				Out: false,
			},
		})
	}

	if resource.Spec.DefaultBackend != nil {
		if resource.Spec.DefaultBackend.Service != nil {
			// +overmind:link Service
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "Service",
					Method: sdp.QueryMethod_GET,
					Query:  resource.Spec.DefaultBackend.Service.Name,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the service affects the ingress' endpoints
					In: true,
					// Changing an ingress does not affect the service
					Out: false,
				},
			})
		}

		if linkRes := resource.Spec.DefaultBackend.Resource; linkRes != nil {
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   linkRes.Kind,
					Method: sdp.QueryMethod_GET,
					Query:  linkRes.Name,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the default backend won't affect the ingress
					// itself
					In: false,
					// Changes to the ingress could affect the default backend
					Out: true,
				},
			})
		}
	}

	for _, rule := range resource.Spec.Rules {
		if rule.Host != "" {
			// +overmind:link dns
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_SEARCH,
					Query:  rule.Host,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Always propagate through rules
					In:  true,
					Out: true,
				},
			})
		}

		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				if path.Backend.Service != nil {
					// +overmind:link Service
					queries = append(queries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "Service",
							Method: sdp.QueryMethod_GET,
							Query:  path.Backend.Service.Name,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changes to the service affects the ingress' endpoints
							In: true,
							// Changing an ingress does not affect the service
							Out: false,
						},
					})
				}

				if path.Backend.Resource != nil {
					queries = append(queries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   path.Backend.Resource.Kind,
							Method: sdp.QueryMethod_GET,
							Query:  path.Backend.Resource.Name,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changes can go in both directions here. An
							// backend change can affect the ingress by rendering
							// backend change can affect the ingress by rending
							// it broken
							In:  true,
							Out: true,
						},
					})
				}
			}
		}
	}

	return queries, nil
}

//go:generate docgen ../docs-data
// +overmind:type Ingress
// +overmind:descriptiveType Ingress
// +overmind:get Get an ingress by name
// +overmind:list List all ingresses
// +overmind:search Search for an ingress using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_ingress_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}

func newIngressSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.Ingress, *v1.IngressList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "Ingress",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.Ingress, *v1.IngressList] {
			return cs.NetworkingV1().Ingresses(namespace)
		},
		ListExtractor: func(list *v1.IngressList) ([]*v1.Ingress, error) {
			extracted := make([]*v1.Ingress, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: ingressExtractor,
	}
}

func init() {
	registerSourceLoader(newIngressSource)
}
