package sources

import (
	v1 "k8s.io/api/rbac/v1"

	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func clusterRoleBindingExtractor(resource *v1.ClusterRoleBinding, scope string) ([]*sdp.Query, error) {
	queries := make([]*sdp.Query, 0)

	queries = append(queries, &sdp.Query{
		Scope:  scope,
		Method: sdp.QueryMethod_GET,
		Query:  resource.RoleRef.Name,
		Type:   resource.RoleRef.Kind,
	})

	for _, subject := range resource.Subjects {
		sd := ScopeDetails{
			ClusterName: scope, // Since this is a cluster role binding, the scope is the cluster name
		}

		if subject.Namespace != "" {
			sd.Namespace = subject.Namespace
		}

		queries = append(queries, &sdp.Query{
			Scope:  sd.String(),
			Method: sdp.QueryMethod_GET,
			Query:  subject.Name,
			Type:   subject.Kind,
		})
	}

	return queries, nil
}

func NewClusterRoleBindingSource(cs *kubernetes.Clientset, cluster string, namespaces []string) KubeTypeSource[*v1.ClusterRoleBinding, *v1.ClusterRoleBindingList] {
	return KubeTypeSource[*v1.ClusterRoleBinding, *v1.ClusterRoleBindingList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "ClusterRoleBinding",
		ClusterInterfaceBuilder: func() ItemInterface[*v1.ClusterRoleBinding, *v1.ClusterRoleBindingList] {
			return cs.RbacV1().ClusterRoleBindings()
		},
		ListExtractor: func(list *v1.ClusterRoleBindingList) ([]*v1.ClusterRoleBinding, error) {
			bindings := make([]*v1.ClusterRoleBinding, len(list.Items))

			for i := range list.Items {
				bindings[i] = &list.Items[i]
			}

			return bindings, nil
		},
		LinkedItemQueryExtractor: clusterRoleBindingExtractor,
	}
}