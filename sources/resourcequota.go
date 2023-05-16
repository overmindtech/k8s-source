package sources

import (
	"github.com/overmindtech/discovery"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

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
