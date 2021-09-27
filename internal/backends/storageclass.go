package backends

import (
	"fmt"

	storageV1 "k8s.io/api/storage/v1"

	"github.com/dylanratcliffe/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// StorageClassSource returns a ResourceSource for StorageClassClaims for a given
// client
func StorageClassSource(cs *kubernetes.Clientset, nss *NamespaceStorage) ResourceSource {
	source := ResourceSource{
		ItemType: "storageclass",
		MapGet:   MapStorageClassGet,
		MapList:  MapStorageClassList,
	}

	source.LoadFunctions(
		cs.StorageV1().StorageClasses().Get,
		cs.StorageV1().StorageClasses().List,
	)

	return source
}

// MapStorageClassList maps an interface that is underneath a
// *storageV1.StorageClassList to a list of Items
func MapStorageClassList(i interface{}) ([]*sdp.Item, error) {
	var objectList *storageV1.StorageClassList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*storageV1.StorageClassList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *storageV1.StorageClassList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapStorageClassGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapStorageClassGet maps an interface that is underneath a *storageV1.StorageClass to an item. If
// the interface isn't actually a *storageV1.StorageClass this will fail
func MapStorageClassGet(i interface{}) (*sdp.Item, error) {
	var object *storageV1.StorageClass
	var item *sdp.Item
	var ok bool

	// Expect this to be a *storageV1.StorageClass
	if object, ok = i.(*storageV1.StorageClass); !ok {
		return item, fmt.Errorf("could not assert %v as a *storageV1.StorageClass", i)
	}

	item, err := mapK8sObject("storageclass", object)

	if err != nil {
		return item, err
	}

	return item, nil
}
