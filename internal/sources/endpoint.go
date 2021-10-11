package sources

import (
	"fmt"

	"github.com/dylanratcliffe/sdp-go"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// EndpointSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func EndpointSource(cs *kubernetes.Clientset, namespace string) ResourceSource {
	source := ResourceSource{
		ItemType: "endpoint",
		MapGet:   MapEndpointGet,
		MapList:  MapEndpointList,
	}

	source.LoadFunctions(
		cs.CoreV1().Endpoints(namespace).Get,
		cs.CoreV1().Endpoints(namespace).List,
	)

	return source
}

// MapEndpointList maps an interface that is underneath a
// *coreV1.EndpointsList to a list of Items
func MapEndpointList(i interface{}) ([]*sdp.Item, error) {
	var objectList *coreV1.EndpointsList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*coreV1.EndpointsList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *coreV1.EndpointsList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapEndpointGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapEndpointGet maps an interface that is underneath a *coreV1.Endpoints to an item. If
// the interface isn't actually a *coreV1.Endpoints this will fail
func MapEndpointGet(i interface{}) (*sdp.Item, error) {
	var object *coreV1.Endpoints
	var ok bool

	// Expect this to be a *coreV1.Endpoints
	if object, ok = i.(*coreV1.Endpoints); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *coreV1.Endpoints", i)
	}

	item, err := mapK8sObject("endpoint", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	// Create linked item requests for all ObjectReferences
	for _, subset := range object.Subsets {
		for _, address := range subset.Addresses {
			if address.TargetRef != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, ObjectReferenceToLIR(address.TargetRef, ClusterName))
			}
		}

		for _, notReadAddress := range subset.NotReadyAddresses {
			if notReadAddress.TargetRef != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, ObjectReferenceToLIR(notReadAddress.TargetRef, ClusterName))
			}
		}
	}

	return item, nil
}
