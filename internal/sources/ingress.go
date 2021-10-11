package sources

import (
	"fmt"

	networkingV1 "k8s.io/api/networking/v1"

	"github.com/dylanratcliffe/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// IngressSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func IngressSource(cs *kubernetes.Clientset, namespace string) ResourceSource {
	source := ResourceSource{
		ItemType: "ingress",
		MapGet:   MapIngressGet,
		MapList:  MapIngressList,
	}

	source.LoadFunctions(
		cs.NetworkingV1().Ingresses(namespace).Get,
		cs.NetworkingV1().Ingresses(namespace).List,
	)

	return source
}

// MapIngressList maps an interface that is underneath a
// *networkingV1.IngressList to a list of Items
func MapIngressList(i interface{}) ([]*sdp.Item, error) {
	var objectList *networkingV1.IngressList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*networkingV1.IngressList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *networkingV1.IngressList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapIngressGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapIngressGet maps an interface that is underneath a *networkingV1.Ingress to an item. If
// the interface isn't actually a *networkingV1.Ingress this will fail
func MapIngressGet(i interface{}) (*sdp.Item, error) {
	var object *networkingV1.Ingress
	var ok bool

	// Expect this to be a *networkingV1.Ingress
	if object, ok = i.(*networkingV1.Ingress); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *networkingV1.Ingress", i)
	}

	item, err := mapK8sObject("ingress", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	// Link services from each path
	for _, rule := range object.Spec.Rules {
		if http := rule.HTTP; http != nil {
			for _, path := range http.Paths {
				if service := path.Backend.Service; service != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Context: item.Context,
						Method:  sdp.RequestMethod_GET,
						Query:   service.Name,
						Type:    "service",
					})
				}
			}
		}
	}

	// Link default if it exists
	if db := object.Spec.DefaultBackend; db != nil {
		if service := db.Service; service != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Context: item.Context,
				Method:  sdp.RequestMethod_GET,
				Query:   service.Name,
				Type:    "service",
			})
		}
	}

	return item, nil
}
