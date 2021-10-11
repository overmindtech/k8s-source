package sources

import (
	"fmt"

	coreV1 "k8s.io/api/scheduling/v1"

	"github.com/dylanratcliffe/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// PriorityClassSource returns a ResourceSource for PriorityClassClaims for a given
// client
func PriorityClassSource(cs *kubernetes.Clientset, nss *NamespaceStorage) ResourceSource {
	source := ResourceSource{
		ItemType: "priorityclass",
		MapGet:   MapPriorityClassGet,
		MapList:  MapPriorityClassList,
	}

	source.LoadFunctions(
		cs.SchedulingV1().PriorityClasses().Get,
		cs.SchedulingV1().PriorityClasses().List,
	)

	return source
}

// MapPriorityClassList maps an interface that is underneath a
// *coreV1.PriorityClassList to a list of Items
func MapPriorityClassList(i interface{}) ([]*sdp.Item, error) {
	var objectList *coreV1.PriorityClassList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*coreV1.PriorityClassList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *coreV1.ClusterRoleBindingList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapPriorityClassGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapPriorityClassGet maps an interface that is underneath a *coreV1.PriorityClass to an item. If
// the interface isn't actually a *coreV1.PriorityClass this will fail
func MapPriorityClassGet(i interface{}) (*sdp.Item, error) {
	var object *coreV1.PriorityClass
	var ok bool

	// Expect this to be a *coreV1.PriorityClass
	if object, ok = i.(*coreV1.PriorityClass); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *coreV1.PriorityClass", i)
	}

	item, err := mapK8sObject("priorityclass", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	return item, nil
}
