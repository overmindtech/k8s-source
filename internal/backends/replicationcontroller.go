package backends

import (
	"fmt"

	"github.com/dylanratcliffe/sdp-go"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ReplicationControllerSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func ReplicationControllerSource(cs *kubernetes.Clientset, namespace string) ResourceSource {
	source := ResourceSource{
		ItemType: "replicationcontroller",
		MapGet:   MapReplicationControllerGet,
		MapList:  MapReplicationControllerList,
	}

	source.LoadFunctions(
		cs.CoreV1().ReplicationControllers(namespace).Get,
		cs.CoreV1().ReplicationControllers(namespace).List,
	)

	return source
}

// MapReplicationControllerList maps an interface that is underneath a
// *coreV1.ReplicationControllerList to a list of Items
func MapReplicationControllerList(i interface{}) ([]*sdp.Item, error) {
	var objectList *coreV1.ReplicationControllerList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*coreV1.ReplicationControllerList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *coreV1.objectList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapReplicationControllerGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapReplicationControllerGet maps an interface that is underneath a *coreV1.ReplicationController to an item. If
// the interface isn't actually a *coreV1.ReplicationController this will fail
func MapReplicationControllerGet(i interface{}) (*sdp.Item, error) {
	var object *coreV1.ReplicationController
	var ok bool

	// Expect this to be a *coreV1.ReplicationController
	if object, ok = i.(*coreV1.ReplicationController); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *coreV1.ReplicationController", i)
	}

	item, err := mapK8sObject("replicationcontroller", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	if object.Spec.Selector != nil {
		item.LinkedItemRequests = []*sdp.ItemRequest{
			// Replication controllers are linked to pods via their selector
			{
				Context: item.Context,
				Method:  sdp.RequestMethod_SEARCH,
				Query: LabelSelectorToQuery(&metaV1.LabelSelector{
					MatchLabels: object.Spec.Selector,
				}),
				Type: "pod",
			},
		}
	}

	return item, nil
}
