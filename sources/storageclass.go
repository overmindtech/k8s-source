package sources

import (
	"github.com/overmindtech/discovery"
	v1 "k8s.io/api/storage/v1"

	"k8s.io/client-go/kubernetes"
)

//go:generate docgen ../docs-data
// +overmind:type StorageClass
// +overmind:descriptiveType Storage Class
// +overmind:get Get a storage class by name
// +overmind:list List all storage classes
// +overmind:search Search for a storage class using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_storage_class.metadata.name
// +overmind:terraform:queryMap kubernetes_storage_class_v1.metadata.name
// +overmind:terraform:scope ${outputs.overmind_kubernetes_cluster_name}.${values.metadata.namespace}

func newStorageClassSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.StorageClass, *v1.StorageClassList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "StorageClass",
		ClusterInterfaceBuilder: func() ItemInterface[*v1.StorageClass, *v1.StorageClassList] {
			return cs.StorageV1().StorageClasses()
		},
		ListExtractor: func(list *v1.StorageClassList) ([]*v1.StorageClass, error) {
			extracted := make([]*v1.StorageClass, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
	}
}

func init() {
	registerSourceLoader(newStorageClassSource)
}
