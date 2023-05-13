package sources

import (
	v1 "k8s.io/api/batch/v1"

	"k8s.io/client-go/kubernetes"
)

func NewCronJobSource(cs *kubernetes.Clientset, cluster string, namespaces []string) KubeTypeSource[*v1.CronJob, *v1.CronJobList] {
	return KubeTypeSource[*v1.CronJob, *v1.CronJobList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "CronJob",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.CronJob, *v1.CronJobList] {
			return cs.BatchV1().CronJobs(namespace)
		},
		ListExtractor: func(list *v1.CronJobList) ([]*v1.CronJob, error) {
			bindings := make([]*v1.CronJob, len(list.Items))

			for i := range list.Items {
				bindings[i] = &list.Items[i]
			}

			return bindings, nil
		},
		// Cronjobs don't need linked items as the jobs they produce are linked
		// automatically
	}
}
