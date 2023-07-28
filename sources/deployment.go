package sources

import (
	"github.com/overmindtech/discovery"
	v1 "k8s.io/api/apps/v1"

	"k8s.io/client-go/kubernetes"
)

//go:generate docgen ../docs-data
// +overmind:type Deployment
// +overmind:descriptiveType Deployment
// +overmind:get Get a deployment by name
// +overmind:list List all deployments
// +overmind:search Search for a deployment using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_deployment.metadata[0].name
// +overmind:terraform:queryMap kubernetes_deployment_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}
// +overmind:link ReplicaSet

func newDeploymentSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.Deployment, *v1.DeploymentList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "Deployment",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.Deployment, *v1.DeploymentList] {
			return cs.AppsV1().Deployments(namespace)
		},
		ListExtractor: func(list *v1.DeploymentList) ([]*v1.Deployment, error) {
			extracted := make([]*v1.Deployment, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		// Replicasets are linked automatically
	}
}

func init() {
	registerSourceLoader(newDeploymentSource)
}
