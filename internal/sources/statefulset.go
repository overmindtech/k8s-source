package sources

import (
	"fmt"

	appsV1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// StatefulSetSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func StatefulSetSource(cs *kubernetes.Clientset) (ResourceSource, error) {
	source := ResourceSource{
		ItemType:   "statefulset",
		MapGet:     MapStatefulSetGet,
		MapList:    MapStatefulSetList,
		Namespaced: true,
	}

	err := source.LoadFunction(
		cs.AppsV1().StatefulSets,
	)

	return source, err
}

// MapStatefulSetList maps an interface that is underneath a
// *appsV1.StatefulSetList to a list of Items
func MapStatefulSetList(i interface{}) ([]*sdp.Item, error) {
	var objectList *appsV1.StatefulSetList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*appsV1.StatefulSetList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *appsV1.StatefulSetList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapStatefulSetGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapStatefulSetGet maps an interface that is underneath a *appsV1.StatefulSet to an item. If
// the interface isn't actually a *appsV1.StatefulSet this will fail
func MapStatefulSetGet(i interface{}) (*sdp.Item, error) {
	var object *appsV1.StatefulSet
	var ok bool

	// Expect this to be a *appsV1.StatefulSet
	if object, ok = i.(*appsV1.StatefulSet); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *appsV1.StatefulSet", i)
	}

	item, err := mapK8sObject("statefulset", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	item.LinkedItemRequests = make([]*sdp.ItemRequest, 0)

	if object.Spec.Selector != nil {
		// Stateful sets are linked to pods via their selector
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Context: item.Context,
			Method:  sdp.RequestMethod_SEARCH,
			Query:   LabelSelectorToQuery(object.Spec.Selector),
			Type:    "pod",
		})
	}

	if object.Spec.ServiceName != "" {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Context: item.Context,
			Method:  sdp.RequestMethod_SEARCH,
			Query: ListOptionsToQuery(&metaV1.ListOptions{
				FieldSelector: Selector{
					"metadata.name":      object.Spec.ServiceName,
					"metadata.namespace": object.Namespace,
				}.String(),
			}),
			Type: "service",
		})
	}

	return item, nil
}
