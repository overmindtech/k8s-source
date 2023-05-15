package sources

import (
	v1 "k8s.io/api/apps/v1"

	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func replicaSetExtractor(resource *v1.ReplicaSet, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	if resource.Spec.Selector != nil {
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Scope:  scope,
				Method: sdp.QueryMethod_SEARCH,
				Query:  LabelSelectorToQuery(resource.Spec.Selector),
				Type:   "Pod",
			},
		})
	}

	return queries, nil
}

func newReplicaSetSource(cs *kubernetes.Clientset, cluster string, namespaces []string) *KubeTypeSource[*v1.ReplicaSet, *v1.ReplicaSetList] {
	return &KubeTypeSource[*v1.ReplicaSet, *v1.ReplicaSetList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "ReplicaSet",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.ReplicaSet, *v1.ReplicaSetList] {
			return cs.AppsV1().ReplicaSets(namespace)
		},
		ListExtractor: func(list *v1.ReplicaSetList) ([]*v1.ReplicaSet, error) {
			extracted := make([]*v1.ReplicaSet, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: replicaSetExtractor,
	}
}
