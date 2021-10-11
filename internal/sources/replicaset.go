package sources

import (
	"fmt"

	appsV1 "k8s.io/api/apps/v1"

	"github.com/dylanratcliffe/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// ReplicaSetSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func ReplicaSetSource(cs *kubernetes.Clientset, namespace string) ResourceSource {
	source := ResourceSource{
		ItemType: "replicaset",
		MapGet:   MapReplicaSetGet,
		MapList:  MapReplicaSetList,
	}

	source.LoadFunctions(
		cs.AppsV1().ReplicaSets(namespace).Get,
		cs.AppsV1().ReplicaSets(namespace).List,
	)

	return source
}

// MapReplicaSetList maps an interface that is underneath a
// *appsV1.ReplicaSetList to a list of Items
func MapReplicaSetList(i interface{}) ([]*sdp.Item, error) {
	var objectList *appsV1.ReplicaSetList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*appsV1.ReplicaSetList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *appsV1.ReplicaSetList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapReplicaSetGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapReplicaSetGet maps an interface that is underneath a *appsV1.ReplicaSet to an item. If
// the interface isn't actually a *appsV1.ReplicaSet this will fail
func MapReplicaSetGet(i interface{}) (*sdp.Item, error) {
	var object *appsV1.ReplicaSet
	var ok bool

	// Expect this to be a *appsV1.ReplicaSet
	if object, ok = i.(*appsV1.ReplicaSet); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *appsV1.ReplicaSet", i)
	}

	item, err := mapK8sObject("replicaset", object)

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
