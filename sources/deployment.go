package sources

import (
	v1 "k8s.io/api/apps/v1"

	"k8s.io/client-go/kubernetes"
)

func newDeploymentSource(cs *kubernetes.Clientset, cluster string, namespaces []string) *KubeTypeSource[*v1.Deployment, *v1.DeploymentList] {
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
