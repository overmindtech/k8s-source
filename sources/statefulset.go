package sources

import (
	v1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func statefulSetExtractor(resource *v1.StatefulSet, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	if resource.Spec.Selector != nil {
		// Stateful sets are linked to pods via their selector
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "Pod",
				Method: sdp.QueryMethod_SEARCH,
				Query:  LabelSelectorToQuery(resource.Spec.Selector),
				Scope:  scope,
			},
		})

		if len(resource.Spec.VolumeClaimTemplates) > 0 {
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "PersistentVolumeClaim",
					Method: sdp.QueryMethod_SEARCH,
					Query:  LabelSelectorToQuery(resource.Spec.Selector),
					Scope:  scope,
				},
			})
		}
	}

	if resource.Spec.ServiceName != "" {
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

func newStatefulSetSource(cs *kubernetes.Clientset, cluster string, namespaces []string) *KubeTypeSource[*v1.StatefulSet, *v1.StatefulSetList] {
	return &KubeTypeSource[*v1.StatefulSet, *v1.StatefulSetList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "StatefulSet",
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
	}
}
