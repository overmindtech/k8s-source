package sources

import (
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func serviceAccountExtractor(resource *v1.ServiceAccount, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	for _, secret := range resource.Secrets {
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Scope:  scope,
				Method: sdp.QueryMethod_GET,
				Query:  secret.Name,
				Type:   "Secret",
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changing the secret will affect the service account and the
				// things that use it
				In: true,
				// The service account cannot affect the secret
				Out: false,
			},
		})
	}

	for _, ipSecret := range resource.ImagePullSecrets {
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Scope:  scope,
				Method: sdp.QueryMethod_GET,
				Query:  ipSecret.Name,
				Type:   "Secret",
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changing the secret will affect the service account and the
				// things that use it
				In: true,
				// The service account cannot affect the secret
				Out: false,
			},
		})
	}

	return queries, nil
}

func newServiceAccountSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.ServiceAccount, *v1.ServiceAccountList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "ServiceAccount",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.ServiceAccount, *v1.ServiceAccountList] {
			return cs.CoreV1().ServiceAccounts(namespace)
		},
		ListExtractor: func(list *v1.ServiceAccountList) ([]*v1.ServiceAccount, error) {
			extracted := make([]*v1.ServiceAccount, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: serviceAccountExtractor,
	}
}

func init() {
	registerSourceLoader(newServiceAccountSource)
}
