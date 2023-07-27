package sources

import (
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func replicationControllerExtractor(resource *v1.ReplicationController, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	if resource.Spec.Selector != nil {
		// +overmind:link Pod
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Scope:  scope,
				Method: sdp.QueryMethod_SEARCH,
				Query: LabelSelectorToQuery(&metaV1.LabelSelector{
					MatchLabels: resource.Spec.Selector,
				}),
				Type: "Pod",
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
// +overmind:type ReplicationController
// +overmind:descriptiveType Replication Controller
// +overmind:get Get a replication controller by name
// +overmind:list List all replication controllers
// +overmind:search Search for a replication controller using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_replication_controller.metadata.name
// +overmind:terraform:queryMap kubernetes_replication_controller_v1.metadata.name
// +overmind:terraform:scope ${outputs.overmind_kubernetes_cluster_name}.${values.metadata.namespace}

func newReplicationControllerSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.ReplicationController, *v1.ReplicationControllerList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "ReplicationController",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.ReplicationController, *v1.ReplicationControllerList] {
			return cs.CoreV1().ReplicationControllers(namespace)
		},
		ListExtractor: func(list *v1.ReplicationControllerList) ([]*v1.ReplicationController, error) {
			extracted := make([]*v1.ReplicationController, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: replicationControllerExtractor,
	}
}

func init() {
	registerSourceLoader(newReplicationControllerSource)
}
