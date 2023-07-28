package sources

import (
	"github.com/overmindtech/discovery"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

//go:generate docgen ../docs-data
// +overmind:type ResourceQuota
// +overmind:descriptiveType Resource Quota
// +overmind:get Get a resource quota by name
// +overmind:list List all resource quotas
// +overmind:search Search for a resource quota using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_resource_quota.metadata.name
// +overmind:terraform:queryMap kubernetes_resource_quota_v1.metadata.name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata.namespace}

func newResourceQuotaSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.ResourceQuota, *v1.ResourceQuotaList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "ResourceQuota",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.ResourceQuota, *v1.ResourceQuotaList] {
			return cs.CoreV1().ResourceQuotas(namespace)
		},
		ListExtractor: func(list *v1.ResourceQuotaList) ([]*v1.ResourceQuota, error) {
			extracted := make([]*v1.ResourceQuota, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
	}
}

func init() {
	registerSourceLoader(newResourceQuotaSource)
}
