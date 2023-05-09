package sources

import (
	"fmt"
	"strings"

	discoveryV1beta1 "k8s.io/api/discovery/v1beta1"

	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// EndpointSliceSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func EndpointSliceSource(cs *kubernetes.Clientset) (ResourceSource, error) {
	source := ResourceSource{
		ItemType:   "endpointslice",
		MapGet:     MapEndpointSliceGet,
		MapList:    MapEndpointSliceList,
		Namespaced: true,
	}

	err := source.LoadFunction(
		cs.DiscoveryV1beta1().EndpointSlices,
	)

	return source, err
}

// MapEndpointSliceList maps an interface that is underneath a
// *discoveryV1beta1.EndpointSliceList to a list of Items
func MapEndpointSliceList(i interface{}) ([]*sdp.Item, error) {
	var objectList *discoveryV1beta1.EndpointSliceList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*discoveryV1beta1.EndpointSliceList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *discoveryV1beta1.EndpointSliceList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapEndpointSliceGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapEndpointSliceGet maps an interface that is underneath a *discoveryV1beta1.EndpointSlice to an item. If
// the interface isn't actually a *discoveryV1beta1.EndpointSlice this will fail
func MapEndpointSliceGet(i interface{}) (*sdp.Item, error) {
	var object *discoveryV1beta1.EndpointSlice
	var ok bool

	// Expect this to be a *discoveryV1beta1.EndpointSlice
	if object, ok = i.(*discoveryV1beta1.EndpointSlice); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *discoveryV1beta1.EndpointSlice", i)
	}

	item, err := mapK8sObject("endpointslice", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	for _, endpoint := range object.Endpoints {
		if tr := endpoint.TargetRef; tr != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Scope:  item.Scope,
				Method: sdp.QueryMethod_GET,
				Query:  tr.Name,
				Type:   strings.ToLower(tr.Kind),
			})
		}
	}

	return item, nil
}
