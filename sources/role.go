package sources

import (
	"github.com/overmindtech/discovery"
	v1 "k8s.io/api/rbac/v1"

	"k8s.io/client-go/kubernetes"
)

//go:generate docgen ../docs-data
// +overmind:type Role
// +overmind:descriptiveType Role
// +overmind:get Get a role by name
// +overmind:list List all roles
// +overmind:search Search for a role using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_role.metadata.name
// +overmind:terraform:queryMap kubernetes_role_v1.metadata.name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata.namespace}

func newRoleSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.Role, *v1.RoleList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "Role",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.Role, *v1.RoleList] {
			return cs.RbacV1().Roles(namespace)
		},
		ListExtractor: func(list *v1.RoleList) ([]*v1.Role, error) {
			extracted := make([]*v1.Role, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
	}
}

func init() {
	registerSourceLoader(newRoleSource)
}
