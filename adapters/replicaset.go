package adapters

import (
	v1 "k8s.io/api/apps/v1"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func replicaSetExtractor(resource *v1.ReplicaSet, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	if resource.Spec.Selector != nil {
		// +overmind:link Pod
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Scope:  scope,
				Method: sdp.QueryMethod_SEARCH,
				Query:  LabelSelectorToQuery(resource.Spec.Selector),
				Type:   "Pod",
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Bidirectional propagation since we control the pods, and the
				// pods host the service
				In:  true,
				Out: true,
			},
		})
	}

	return queries, nil
}

//go:generate docgen ../docs-data
// +overmind:type ReplicaSet
// +overmind:descriptiveType Replica Set
// +overmind:get Get a replica set by name
// +overmind:list List all replica sets
// +overmind:search Search for a replica set using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes

func newReplicaSetAdapter(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Adapter {
	return &KubeTypeAdapter[*v1.ReplicaSet, *v1.ReplicaSetList]{
		ClusterName:      cluster,
		Namespaces:       namespaces,
		TypeName:         "ReplicaSet",
		AutoQueryExtract: true,
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

func init() {
	registerAdapterLoader(newReplicaSetAdapter)
}
