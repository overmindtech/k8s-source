package sources

import (
	v1 "k8s.io/api/batch/v1"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func jobExtractor(resource *v1.Job, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	if resource.Spec.Selector != nil {
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Scope:  scope,
				Method: sdp.QueryMethod_SEARCH,
				Query:  LabelSelectorToQuery(resource.Spec.Selector),
				Type:   "Pod",
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changes to a job will replace the pods, changes to the pods
				// could break the job
				In:  true,
				Out: true,
			},
		})
	}

	return queries, nil
}

func newJobSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.Job, *v1.JobList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "Job",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.Job, *v1.JobList] {
			return cs.BatchV1().Jobs(namespace)
		},
		ListExtractor: func(list *v1.JobList) ([]*v1.Job, error) {
			extracted := make([]*v1.Job, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: jobExtractor,
	}
}

func init() {
	registerSourceLoader(newJobSource)
}
