package adapters

import (
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/scheduling/v1"

	"k8s.io/client-go/kubernetes"
)

//go:generate docgen ../docs-data
// +overmind:type PriorityClass
// +overmind:descriptiveType Priority Class
// +overmind:get Get a priority class by name
// +overmind:list List all priority classes
// +overmind:search Search for a THING using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_priority_class.metadata[0].name
// +overmind:terraform:queryMap kubernetes_priority_class_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}

func newPriorityClassAdapter(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Adapter {
	return &KubeTypeAdapter[*v1.PriorityClass, *v1.PriorityClassList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "PriorityClass",
		ClusterInterfaceBuilder: func() ItemInterface[*v1.PriorityClass, *v1.PriorityClassList] {
			return cs.SchedulingV1().PriorityClasses()
		},
		ListExtractor: func(list *v1.PriorityClassList) ([]*v1.PriorityClass, error) {
			extracted := make([]*v1.PriorityClass, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		AdapterMetadata: priorityClassAdapterMetadata,
	}
}

var priorityClassAdapterMetadata = AdapterMetadata.Register(&sdp.AdapterMetadata{
	Type:                  "PriorityClass",
	DescriptiveName:       "Priority Class",
	Category:              sdp.AdapterCategory_ADAPTER_CATEGORY_CONFIGURATION,
	SupportedQueryMethods: DefaultSupportedQueryMethods("Priority Class"),
	TerraformMappings: []*sdp.TerraformMapping{
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_priority_class_v1.metadata[0].name",
		},
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_priority_class.metadata[0].name",
		},
	},
})

func init() {
	registerAdapterLoader(newPriorityClassAdapter)
}
