package sources

import (
	v1 "k8s.io/api/rbac/v1"

	"k8s.io/client-go/kubernetes"
)

func newRoleSource(cs *kubernetes.Clientset, cluster string, namespaces []string) *KubeTypeSource[*v1.Role, *v1.RoleList] {
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
