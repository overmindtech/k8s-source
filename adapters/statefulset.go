package adapters

import (
	v1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func statefulSetExtractor(resource *v1.StatefulSet, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	if resource.Spec.Selector != nil {
		// Stateful sets are linked to pods via their selector
		// +overmind:link Pod
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "Pod",
				Method: sdp.QueryMethod_SEARCH,
				Query:  LabelSelectorToQuery(resource.Spec.Selector),
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Bidirectional propagation since we control the pods, and the
				// pods host the stateful set
				In:  true,
				Out: true,
			},
		})

		if len(resource.Spec.VolumeClaimTemplates) > 0 {
			// +overmind:link PersistentVolumeClaim
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "PersistentVolumeClaim",
					Method: sdp.QueryMethod_SEARCH,
					Query:  LabelSelectorToQuery(resource.Spec.Selector),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Bidirectional propagation since we control the pods, and the
					// pods host the stateful set
					In:  true,
					Out: true,
				},
			})
		}
	}

	if resource.Spec.ServiceName != "" {
		// +overmind:link Service
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Scope:  scope,
				Method: sdp.QueryMethod_SEARCH,
				Query: ListOptionsToQuery(&metaV1.ListOptions{
					FieldSelector: Selector{
						"metadata.name":      resource.Spec.ServiceName,
						"metadata.namespace": resource.Namespace,
					}.String(),
				}),
				Type: "Service",
			},
		})
	}

	return queries, nil
}

//go:generate docgen ../docs-data
// +overmind:type StatefulSet
// +overmind:descriptiveType Stateful Set
// +overmind:get Get a stateful set by name
// +overmind:list List all stateful sets
// +overmind:search Search for a stateful set using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_stateful_set.metadata[0].name
// +overmind:terraform:queryMap kubernetes_stateful_set_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}

func newStatefulSetAdapter(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Adapter {
	return &KubeTypeAdapter[*v1.StatefulSet, *v1.StatefulSetList]{
		ClusterName:      cluster,
		Namespaces:       namespaces,
		TypeName:         "StatefulSet",
		AutoQueryExtract: true,
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.StatefulSet, *v1.StatefulSetList] {
			return cs.AppsV1().StatefulSets(namespace)
		},
		ListExtractor: func(list *v1.StatefulSetList) ([]*v1.StatefulSet, error) {
			extracted := make([]*v1.StatefulSet, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: statefulSetExtractor,
		AdapterMetadata:          statefulSetAdapterMetadata,
	}
}

var statefulSetAdapterMetadata = Metadata.Register(&sdp.AdapterMetadata{
	Type:                  "StatefulSet",
	Category:              sdp.AdapterCategory_ADAPTER_CATEGORY_COMPUTE_APPLICATION,
	DescriptiveName:       "Stateful Set",
	SupportedQueryMethods: DefaultSupportedQueryMethods("Stateful Set"),
	TerraformMappings: []*sdp.TerraformMapping{
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_stateful_set_v1.metadata[0].name",
		},
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_stateful_set.metadata[0].name",
		},
	},
})

func init() {
	registerAdapterLoader(newStatefulSetAdapter)
}
