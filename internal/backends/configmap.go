package backends

import (
	"fmt"

	"github.com/dylanratcliffe/sdp-go"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// ConfigMapSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func ConfigMapSource(cs *kubernetes.Clientset, namespace string) ResourceSource {
	source := ResourceSource{
		ItemType: "configMap",
		MapGet:   MapConfigMapGet,
		MapList:  MapConfigMapList,
	}

	source.LoadFunctions(
		cs.CoreV1().ConfigMaps(namespace).Get,
		cs.CoreV1().ConfigMaps(namespace).List,
	)

	return source
}

// MapConfigMapList maps an interface that is underneath a
// *coreV1.ConfigMapList to a list of Items
func MapConfigMapList(i interface{}) ([]*sdp.Item, error) {
	var objectList *coreV1.ConfigMapList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*coreV1.ConfigMapList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *coreV1.ConfigMapList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapConfigMapGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapConfigMapGet maps an interface that is underneath a *coreV1.ConfigMap to an item. If
// the interface isn't actually a *coreV1.ConfigMap this will fail
func MapConfigMapGet(i interface{}) (*sdp.Item, error) {
	var object *coreV1.ConfigMap
	var ok bool

	// Expect this to be a *coreV1.ConfigMap
	if object, ok = i.(*coreV1.ConfigMap); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *coreV1.ConfigMap", i)
	}

	item, err := mapK8sObject("configMap", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	return item, nil
}
