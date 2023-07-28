package sources

import (
	"github.com/overmindtech/discovery"
	v1 "k8s.io/api/batch/v1"

	"k8s.io/client-go/kubernetes"
)

//go:generate docgen ../docs-data
// +overmind:type CronJob
// +overmind:descriptiveType Cron Job
// +overmind:get Get a cron job by name
// +overmind:list List all cron jobs
// +overmind:search Search for a cron job using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_cron_job.metadata.name
// +overmind:terraform:queryMap kubernetes_cron_job_v1.metadata.name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}

func newCronJobSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.CronJob, *v1.CronJobList]{
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

func init() {
	registerSourceLoader(newCronJobSource)
}
