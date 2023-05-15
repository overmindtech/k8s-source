package sources

import (
	"strings"

	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
)

func linkedItemExtractor(resource *v1.Node, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	for _, addr := range resource.Status.Addresses {
		switch addr.Type {
		case v1.NodeExternalDNS:
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_GET,
					Query:  addr.Address,
					Scope:  "global",
				},
			})
		case v1.NodeExternalIP, v1.NodeInternalIP:
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ip",
					Method: sdp.QueryMethod_GET,
					Query:  addr.Address,
					Scope:  "global",
				},
			})
		}
	}

	for _, vol := range resource.Status.VolumesAttached {
		// Look for EBS volumes since they follow the format:
		// kubernetes.io/csi/ebs.csi.aws.com^vol-043e04d9cc6d72183
		if strings.HasPrefix(string(vol.Name), "kubernetes.io/csi/ebs.csi.aws.com") {
			sections := strings.Split(string(vol.Name), "^")

			if len(sections) == 2 {
				queries = append(queries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-volume",
						Method: sdp.QueryMethod_GET,
						Query:  sections[1],
						Scope:  "*",
					},
				})
			}
		}
	}

	return queries, nil
}

// TODO: Should we try a DNS lookup for a node name? Is the hostname stored anywhere?

func newNodeSource(cs *kubernetes.Clientset, cluster string, namespaces []string) *KubeTypeSource[*v1.Node, *v1.NodeList] {
	return &KubeTypeSource[*v1.Node, *v1.NodeList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "Node",
		ClusterInterfaceBuilder: func() ItemInterface[*v1.Node, *v1.NodeList] {
			return cs.CoreV1().Nodes()
		},
		ListExtractor: func(list *v1.NodeList) ([]*v1.Node, error) {
			extracted := make([]*v1.Node, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: linkedItemExtractor,
	}
}
