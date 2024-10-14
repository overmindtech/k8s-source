package adapters

import (
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

//go:generate docgen ../docs-data
// +overmind:type ConfigMap
// +overmind:descriptiveType Config Map
// +overmind:get Get a config map by name
// +overmind:list List all config maps
// +overmind:search Search for a config map using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_config_map.metadata[0].name
// +overmind:terraform:queryMap kubernetes_config_map_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}

func newConfigMapAdapter(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Adapter {
	return &KubeTypeAdapter[*v1.ConfigMap, *v1.ConfigMapList]{
		ClusterName:      cluster,
		Namespaces:       namespaces,
		TypeName:         "ConfigMap",
		AutoQueryExtract: true,
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.ConfigMap, *v1.ConfigMapList] {
			return cs.CoreV1().ConfigMaps(namespace)
		},
		ListExtractor: func(list *v1.ConfigMapList) ([]*v1.ConfigMap, error) {
			bindings := make([]*v1.ConfigMap, len(list.Items))

			for i := range list.Items {
				bindings[i] = &list.Items[i]
			}

			return bindings, nil
		},
		AdapterMetadata: configMapAdapterMetadata,
	}
}

var configMapAdapterMetadata = Metadata.Register(&sdp.AdapterMetadata{
	Type:                  "ConfigMap",
	Category:              sdp.AdapterCategory_ADAPTER_CATEGORY_CONFIGURATION,
	DescriptiveName:       "Config Map",
	SupportedQueryMethods: DefaultSupportedQueryMethods("Config Map"),
	TerraformMappings: []*sdp.TerraformMapping{
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_config_map_v1.metadata[0].name",
		},
		{
			TerraformMethod:   sdp.QueryMethod_GET,
			TerraformQueryMap: "kubernetes_config_map.metadata[0].name",
		},
	},
})

func init() {
	registerAdapterLoader(newConfigMapAdapter)
}
