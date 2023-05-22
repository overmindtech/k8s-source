package sources

import (
	v2 "k8s.io/api/autoscaling/v2"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func horizontalPodAutoscalerExtractor(resource *v2.HorizontalPodAutoscaler, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	queries = append(queries, &sdp.LinkedItemQuery{
		Query: &sdp.Query{
			Type:   resource.Spec.ScaleTargetRef.Kind,
			Method: sdp.QueryMethod_GET,
			Query:  resource.Spec.ScaleTargetRef.Name,
			Scope:  scope,
		},
		BlastPropagation: &sdp.BlastPropagation{
			// Changes to the target won't affect the hpa
			In: false,
			// Changes to the hpa can affect the target i.e. by scaling the pods
			Out: true,
		},
	})

	return queries, nil
}

func newHorizontalPodAutoscalerSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v2.HorizontalPodAutoscaler, *v2.HorizontalPodAutoscalerList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "HorizontalPodAutoscaler",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v2.HorizontalPodAutoscaler, *v2.HorizontalPodAutoscalerList] {
			return cs.AutoscalingV2().HorizontalPodAutoscalers(namespace)
		},
		ListExtractor: func(list *v2.HorizontalPodAutoscalerList) ([]*v2.HorizontalPodAutoscaler, error) {
			extracted := make([]*v2.HorizontalPodAutoscaler, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: horizontalPodAutoscalerExtractor,
	}
}

func init() {
	registerSourceLoader(newHorizontalPodAutoscalerSource)
}
