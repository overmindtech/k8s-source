package sources

import (
	v1 "k8s.io/api/rbac/v1"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func roleBindingExtractor(resource *v1.RoleBinding, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	sd, err := ParseScope(scope, true)

	if err != nil {
		return nil, err
	}

	for _, subject := range resource.Subjects {
		// +overmind:link ServiceAccount
		// +overmind:link User
		// +overmind:link Group
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Method: sdp.QueryMethod_GET,
				Query:  subject.Name,
				Type:   subject.Kind,
				Scope: ScopeDetails{
					ClusterName: sd.ClusterName,
					Namespace:   subject.Namespace,
				}.String(),
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changes to the subject (the group we're applying permissions
				// to) won't affect the role or the binding
				In: false,
				// Changes to the binding will affect the subject
				Out: true,
			},
		})
	}

	refSD := ScopeDetails{
		ClusterName: sd.ClusterName,
	}

	switch resource.RoleRef.Kind {
	case "Role":
		// +overmind:link Role
		// If this binding is linked to a role then it's in the same namespace
		refSD.Namespace = sd.Namespace
	case "ClusterRole":
		// +overmind:link ClusterRole
		// If this is linked to a ClusterRole (which is not namespaced) we need
		// to make sure that we are querying the root scope i.e. the
		// non-namespaced scope
		refSD.Namespace = ""
	}

	queries = append(queries, &sdp.LinkedItemQuery{
		Query: &sdp.Query{
			Scope:  refSD.String(),
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

	return queries, nil
}

//go:generate docgen ../docs-data
// +overmind:type RoleBinding
// +overmind:descriptiveType Role Binding
// +overmind:get Get a role binding by name
// +overmind:list List all role bindings
// +overmind:search Search for a role binding using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_role_binding.metadata[0].name
// +overmind:terraform:queryMap kubernetes_role_binding_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}

func newRoleBindingSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.RoleBinding, *v1.RoleBindingList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "RoleBinding",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.RoleBinding, *v1.RoleBindingList] {
			return cs.RbacV1().RoleBindings(namespace)
		},
		ListExtractor: func(list *v1.RoleBindingList) ([]*v1.RoleBinding, error) {
			extracted := make([]*v1.RoleBinding, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: roleBindingExtractor,
	}
}

func init() {
	registerSourceLoader(newRoleBindingSource)
}
