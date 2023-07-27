package sources

import (
	"github.com/overmindtech/discovery"
	v1 "k8s.io/api/rbac/v1"

	"k8s.io/client-go/kubernetes"
)

//go:generate docgen ../docs-data
// +overmind:type ClusterRole
// +overmind:descriptiveType Cluster Role
// +overmind:get Get a cluster role by name
// +overmind:list List all cluster roles
// +overmind:search Search for a THING using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_cluster_role_v1.metadata.name
// +overmind:terraform:scope ${outputs.overmind_kubernetes_cluster_name}

func newClusterRoleSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.ClusterRole, *v1.ClusterRoleList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "ClusterRole",
		ClusterInterfaceBuilder: func() ItemInterface[*v1.ClusterRole, *v1.ClusterRoleList] {
			return cs.RbacV1().ClusterRoles()
		},
		ListExtractor: func(list *v1.ClusterRoleList) ([]*v1.ClusterRole, error) {
			bindings := make([]*v1.ClusterRole, len(list.Items))

			for i := range list.Items {
				bindings[i] = &list.Items[i]
			}

			return bindings, nil
		},
	}
}

func init() {
	registerSourceLoader(newClusterRoleSource)
}
