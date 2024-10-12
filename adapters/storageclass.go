package adapters

import (
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
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
// +overmind:terraform:queryMap kubernetes_storage_class.metadata[0].name
// +overmind:terraform:queryMap kubernetes_storage_class_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}

func newStorageClassAdapter(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Adapter {
	return &KubeTypeAdapter[*v1.StorageClass, *v1.StorageClassList]{
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
		AdapterMetadata: storageClassAdapterMetadata,
	}
}

var storageClassAdapterMetadata = AdapterMetadata.Register(&sdp.AdapterMetadata{
	Type:                  "StorageClass",
	DescriptiveName:       "Storage Class",
	Category:              sdp.AdapterCategory_ADAPTER_CATEGORY_STORAGE,
	SupportedQueryMethods: DefaultSupportedQueryMethods("Storage Class"),
	TerraformMappings: []*sdp.TerraformMapping{
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_storage_class.metadata[0].name",
		},
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_storage_class_v1.metadata[0].name",
		},
	},
})

func init() {
	registerAdapterLoader(newStorageClassAdapter)
}
