package sources

import (
	"fmt"

	networkingV1 "k8s.io/api/networking/v1"

	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// NetworkPolicySource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func NetworkPolicySource(cs *kubernetes.Clientset) (ResourceSource, error) {
	source := ResourceSource{
		ItemType:   "networkpolicy",
		MapGet:     MapNetworkPolicyGet,
		MapList:    MapNetworkPolicyList,
		Namespaced: true,
	}

	err := source.LoadFunction(
		cs.NetworkingV1().NetworkPolicies,
	)

	return source, err
}

// MapNetworkPolicyList maps an interface that is underneath a
// *networkingV1.NetworkPolicyList to a list of Items
func MapNetworkPolicyList(i interface{}) ([]*sdp.Item, error) {
	var objectList *networkingV1.NetworkPolicyList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*networkingV1.NetworkPolicyList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *networkingV1.NetworkPolicyList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapNetworkPolicyGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapNetworkPolicyGet maps an interface that is underneath a *networkingV1.NetworkPolicy to an item. If
// the interface isn't actually a *networkingV1.NetworkPolicy this will fail
func MapNetworkPolicyGet(i interface{}) (*sdp.Item, error) {
	var object *networkingV1.NetworkPolicy
	var ok bool

	// Expect this to be a *networkingV1.NetworkPolicy
	if object, ok = i.(*networkingV1.NetworkPolicy); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *networkingV1.NetworkPolicy", i)
	}

	item, err := mapK8sObject("networkpolicy", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
		Scope:  item.Scope,
		Method: sdp.QueryMethod_GET,
		Query:  LabelSelectorToQuery(&object.Spec.PodSelector),
		Type:   "pod",
	})

	var peers []networkingV1.NetworkPolicyPeer

	for _, ig := range object.Spec.Ingress {
		peers = append(peers, ig.From...)
	}

	for _, eg := range object.Spec.Egress {
		peers = append(peers, eg.To...)
	}

	// Link all peers
	for _, peer := range peers {
		if ps := peer.PodSelector; ps != nil {
			// TODO: Link to namespaces that are allowed to ingress e.g.
			// - namespaceSelector:
			// matchLabels:
			//   project: something

			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Scope:  item.Scope,
				Method: sdp.QueryMethod_GET,
				Query:  LabelSelectorToQuery(ps),
				Type:   "pod",
			})
		}
	}

	return item, nil
}
