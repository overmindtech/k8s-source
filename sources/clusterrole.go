package sources

import (
	"github.com/overmindtech/discovery"
	v1 "k8s.io/api/rbac/v1"

	"k8s.io/client-go/kubernetes"
)

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