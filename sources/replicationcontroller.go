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
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Scope:  scope,
				Method: sdp.QueryMethod_SEARCH,
				Query: LabelSelectorToQuery(&metaV1.LabelSelector{
					MatchLabels: resource.Spec.Selector,
				}),
				Type: "Pod",
			},
		})
	}

	return queries, nil
}

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
