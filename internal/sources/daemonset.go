package sources

import (
	"fmt"

	appsV1 "k8s.io/api/apps/v1"

	"github.com/dylanratcliffe/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// DaemonSetSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func DaemonSetSource(cs *kubernetes.Clientset) ResourceSource {
	source := ResourceSource{
		ItemType:   "daemonset",
		MapGet:     MapDaemonSetGet,
		MapList:    MapDaemonSetList,
		Namespaced: true,
	}

	source.LoadFunction(
		cs.AppsV1().DaemonSets,
	)

	return source
}

// MapDaemonSetList maps an interface that is underneath a
// *appsV1.DaemonSetList to a list of Items
func MapDaemonSetList(i interface{}) ([]*sdp.Item, error) {
	var objectList *appsV1.DaemonSetList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*appsV1.DaemonSetList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *appsV1.DaemonSetList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapDaemonSetGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapDaemonSetGet maps an interface that is underneath a *appsV1.DaemonSet to an item. If
// the interface isn't actually a *appsV1.DaemonSet this will fail
func MapDaemonSetGet(i interface{}) (*sdp.Item, error) {
	var object *appsV1.DaemonSet
	var ok bool

	// Expect this to be a *appsV1.DaemonSet
	if object, ok = i.(*appsV1.DaemonSet); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *appsV1.DaemonSet", i)
	}

	item, err := mapK8sObject("daemonset", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	if object.Spec.Selector != nil {
		item.LinkedItemRequests = []*sdp.ItemRequest{
			// Services are linked to pods via their selector
			{
				Context: item.Context,
				Method:  sdp.RequestMethod_SEARCH,
				Query:   LabelSelectorToQuery(object.Spec.Selector),
				Type:    "pod",
			},
		}
	}

	return item, nil
}
