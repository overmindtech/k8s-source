package sources

import (
	v1 "k8s.io/api/rbac/v1"

	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

func roleBindingExtractor(resource *v1.RoleBinding, scope string) ([]*sdp.Query, error) {
	queries := make([]*sdp.Query, 0)

	sd, err := ParseScope(scope, true)

	if err != nil {
		return nil, err
	}

	for _, subject := range resource.Subjects {
		queries = append(queries, &sdp.Query{
			Method: sdp.QueryMethod_GET,
			Query:  subject.Name,
			Type:   subject.Kind,
			Scope: ScopeDetails{
				ClusterName: sd.ClusterName,
				Namespace:   subject.Namespace,
			}.String(),
		})
	}

	refSD := ScopeDetails{
		ClusterName: sd.ClusterName,
	}

	switch resource.RoleRef.Kind {
	case "Role":
		// If this binding is linked to a role then it's in the same namespace
		refSD.Namespace = sd.Namespace
	case "ClusterRole":
		// If this is linked to a ClusterRole (which is not namespaced) we need
		// to make sure that we are querying the root scope i.e. the
		// non-namespaced scope
		refSD.Namespace = ""
	}

	queries = append(queries, &sdp.Query{
		Scope:  refSD.String(),
		Method: sdp.QueryMethod_GET,
		Query:  resource.RoleRef.Name,
		Type:   resource.RoleRef.Kind,
	})

	return queries, nil
}

func NewRoleBindingSource(cs *kubernetes.Clientset, cluster string, namespaces []string) KubeTypeSource[*v1.RoleBinding, *v1.RoleBindingList] {
	return KubeTypeSource[*v1.RoleBinding, *v1.RoleBindingList]{
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