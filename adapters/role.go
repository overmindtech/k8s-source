package adapters

import (
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
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
// +overmind:terraform:queryMap kubernetes_role.metadata[0].name
// +overmind:terraform:queryMap kubernetes_role_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}

func newRoleAdapter(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Adapter {
	return &KubeTypeAdapter[*v1.Role, *v1.RoleList]{
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
		AdapterMetadata: roleAdapterMetadata,
	}
}

var roleAdapterMetadata = Metadata.Register(&sdp.AdapterMetadata{
	Type:                  "Role",
	DescriptiveName:       "Role",
	Category:              sdp.AdapterCategory_ADAPTER_CATEGORY_SECURITY,
	SupportedQueryMethods: DefaultSupportedQueryMethods("Role"),
	TerraformMappings: []*sdp.TerraformMapping{
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_role_v1.metadata[0].name",
		},
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_role.metadata[0].name",
		},
	},
})

func init() {
	registerAdapterLoader(newRoleAdapter)
}
