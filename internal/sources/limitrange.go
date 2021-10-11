package sources

import (
	"fmt"

	"github.com/dylanratcliffe/sdp-go"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// LimitRangeSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func LimitRangeSource(cs *kubernetes.Clientset) ResourceSource {
	source := ResourceSource{
		ItemType:   "limitrange",
		MapGet:     MapLimitRangeGet,
		MapList:    MapLimitRangeList,
		Namespaced: true,
	}

	source.LoadFunction(
		cs.CoreV1().LimitRanges,
	)

	return source
}

// MapLimitRangeList maps an interface that is underneath a
// *coreV1.LimitRangeList to a list of Items
func MapLimitRangeList(i interface{}) ([]*sdp.Item, error) {
	var objectList *coreV1.LimitRangeList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*coreV1.LimitRangeList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *coreV1.LimitRangeList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapLimitRangeGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapLimitRangeGet maps an interface that is underneath a *coreV1.LimitRange to an item. If
// the interface isn't actually a *coreV1.LimitRange this will fail
func MapLimitRangeGet(i interface{}) (*sdp.Item, error) {
	var object *coreV1.LimitRange
	var ok bool

	// Expect this to be a *coreV1.LimitRange
	if object, ok = i.(*coreV1.LimitRange); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *coreV1.LimitRange", i)
	}

	item, err := mapK8sObject("limitrange", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	// TODO: Should these be linked to a namespace? The only thing that a limit
	// range is actually related to is the namespace

	return item, nil
}
