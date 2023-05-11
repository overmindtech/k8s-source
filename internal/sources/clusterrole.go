package sources

import (
	v1 "k8s.io/api/rbac/v1"

	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func NewClusterRoleSource(cs *kubernetes.Clientset, cluster string, namespaces []string) KubeTypeSource[*v1.ClusterRole, *v1.ClusterRoleList] {
	return KubeTypeSource[*v1.ClusterRole, *v1.ClusterRoleList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "ClusterRole",
		ClusterInterfaceBuilder: func() ItemInterface[*v1.ClusterRole, *v1.ClusterRoleList] {
			return cs.RbacV1().ClusterRoles()
		},
		ListExtractor: func(list *v1.ClusterRoleList) ([]*v1.ClusterRole, error) {
			bindings := make([]*v1.ClusterRole, len(list.Items))

			for i, cr := range list.Items {
				bindings[i] = &cr
			}

			return bindings, nil
		},
		LinkedItemQueryExtractor: func(resource *v1.ClusterRole, scope string) ([]*sdp.Query, error) {
			// No linked items
			return []*sdp.Query{}, nil
		},
	}
}
