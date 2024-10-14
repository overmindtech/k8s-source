package adapters

import (
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

//go:generate docgen ../docs-data
// +overmind:type LimitRange
// +overmind:descriptiveType Limit Range
// +overmind:get Get a limit range by name
// +overmind:list List all limit ranges
// +overmind:search Search for a limit range using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_limit_range.metadata[0].name
// +overmind:terraform:queryMap kubernetes_limit_range_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}

func newLimitRangeAdapter(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Adapter {
	return &KubeTypeAdapter[*v1.LimitRange, *v1.LimitRangeList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "LimitRange",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.LimitRange, *v1.LimitRangeList] {
			return cs.CoreV1().LimitRanges(namespace)
		},
		ListExtractor: func(list *v1.LimitRangeList) ([]*v1.LimitRange, error) {
			extracted := make([]*v1.LimitRange, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		AdapterMetadata: limitRangeAdapterMetadata,
	}
}

var limitRangeAdapterMetadata = Metadata.Register(&sdp.AdapterMetadata{
	Type:                  "LimitRange",
	DescriptiveName:       "Limit Range",
	Category:              sdp.AdapterCategory_ADAPTER_CATEGORY_CONFIGURATION,
	SupportedQueryMethods: DefaultSupportedQueryMethods("Limit Range"),
	TerraformMappings: []*sdp.TerraformMapping{
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_limit_range_v1.metadata[0].name",
		},
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_limit_range.metadata[0].name",
		},
	},
})

func init() {
	registerAdapterLoader(newLimitRangeAdapter)
}
