package sources

import (
	v1 "k8s.io/api/networking/v1"

	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func ingressExtractor(resource *v1.Ingress, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	if resource.Spec.IngressClassName != nil {
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "IngressClass",
				Method: sdp.QueryMethod_GET,
				Query:  *resource.Spec.IngressClassName,
				Scope:  scope,
			},
		})
	}

	if resource.Spec.DefaultBackend != nil {
		if resource.Spec.DefaultBackend.Service != nil {
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "Service",
					Method: sdp.QueryMethod_GET,
					Query:  resource.Spec.DefaultBackend.Service.Name,
					Scope:  scope,
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
			})
		}
	}

	for _, rule := range resource.Spec.Rules {
		if rule.Host != "" {
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_GET,
					Query:  rule.Host,
					Scope:  "global",
				},
			})
		}

		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				if path.Backend.Service != nil {
					queries = append(queries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "Service",
							Method: sdp.QueryMethod_GET,
							Query:  path.Backend.Service.Name,
							Scope:  scope,
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
					})
				}
			}
		}
	}

	return queries, nil
}

func newIngressSource(cs *kubernetes.Clientset, cluster string, namespaces []string) *KubeTypeSource[*v1.Ingress, *v1.IngressList] {
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
