package adapters

import (
	v1 "k8s.io/api/rbac/v1"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func clusterRoleBindingExtractor(resource *v1.ClusterRoleBinding, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	// +overmind:link ClusterRole
	queries = append(queries, &sdp.LinkedItemQuery{
		Query: &sdp.Query{
			Scope:  scope,
			Method: sdp.QueryMethod_GET,
			Query:  resource.RoleRef.Name,
			Type:   resource.RoleRef.Kind,
		},
		BlastPropagation: &sdp.BlastPropagation{
			// Changes to the role will affect the things bound to it since the
			// role contains the permissions
			In: true,
			// Changes to the binding won't affect the role
			Out: false,
		},
	})

	for _, subject := range resource.Subjects {
		sd := ScopeDetails{
			ClusterName: scope, // Since this is a cluster role binding, the scope is the cluster name
		}

		if subject.Namespace != "" {
			sd.Namespace = subject.Namespace
		}

		// +overmind:link ServiceAccount
		// +overmind:link User
		// +overmind:link Group
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Scope:  sd.String(),
				Method: sdp.QueryMethod_GET,
				Query:  subject.Name,
				Type:   subject.Kind,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changes to the group won't affect the binding itself
				In: false,
				// Changes to the binding will affect the group it's bound to
				Out: true,
			},
		})
	}

	return queries, nil
}

//go:generate docgen ../docs-data
// +overmind:type ClusterRoleBinding
// +overmind:descriptiveType Cluster Role Binding
// +overmind:get Get a cluster role binding by name
// +overmind:list List all cluster role bindings
// +overmind:search Search for a cluster role binding using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_cluster_role_binding.metadata[0].name
// +overmind:terraform:queryMap kubernetes_cluster_role_binding_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}

func newClusterRoleBindingAdapter(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Adapter {
	return &KubeTypeAdapter[*v1.ClusterRoleBinding, *v1.ClusterRoleBindingList]{
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

func init() {
	registerAdapterLoader(newClusterRoleBindingAdapter)
}
