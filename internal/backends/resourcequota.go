package backends

import (
	"fmt"

	"github.com/dylanratcliffe/sdp-go"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// ResourceQuotaSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func ResourceQuotaSource(cs *kubernetes.Clientset, namespace string) ResourceSource {
	source := ResourceSource{
		ItemType: "resourcequota",
		MapGet:   MapResourceQuotaGet,
		MapList:  MapResourceQuotaList,
	}

	source.LoadFunctions(
		cs.CoreV1().ResourceQuotas(namespace).Get,
		cs.CoreV1().ResourceQuotas(namespace).List,
	)

	return source
}

// MapResourceQuotaList maps an interface that is underneath a
// *coreV1.ResourceQuotaList to a list of Items
func MapResourceQuotaList(i interface{}) ([]*sdp.Item, error) {
	var objectList *coreV1.ResourceQuotaList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*coreV1.ResourceQuotaList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *coreV1.ResourceQuotaList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapResourceQuotaGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapResourceQuotaGet maps an interface that is underneath a *coreV1.ResourceQuota to an item. If
// the interface isn't actually a *coreV1.ResourceQuota this will fail
func MapResourceQuotaGet(i interface{}) (*sdp.Item, error) {
	var object *coreV1.ResourceQuota
	var ok bool

	// Expect this to be a *coreV1.ResourceQuota
	if object, ok = i.(*coreV1.ResourceQuota); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *coreV1.ResourceQuota", i)
	}

	item, err := mapK8sObject("resourcequota", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	return item, nil
}
