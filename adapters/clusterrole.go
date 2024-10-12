package adapters

import (
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
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
// +overmind:terraform:queryMap kubernetes_cluster_role_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}

func newClusterRoleAdapter(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Adapter {
	return &KubeTypeAdapter[*v1.ClusterRole, *v1.ClusterRoleList]{
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
		AdapterMetadata: clusterRoleAdapterMetadata,
	}
}

var clusterRoleAdapterMetadata = AdapterMetadata.Register(&sdp.AdapterMetadata{
	Type:                  "ClusterRole",
	Category:              sdp.AdapterCategory_ADAPTER_CATEGORY_SECURITY,
	DescriptiveName:       "Cluster Role",
	SupportedQueryMethods: DefaultSupportedQueryMethods("Cluster Role"),
	TerraformMappings: []*sdp.TerraformMapping{
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_cluster_role_v1.metadata[0].name",
		},
	},
})

func init() {
	registerAdapterLoader(newClusterRoleAdapter)
}
