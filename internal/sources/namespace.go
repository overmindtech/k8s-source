package sources

import (
	"fmt"
	"strings"

	"github.com/dylanratcliffe/sdp-go"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// NamespaceSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func NamespaceSource(cs *kubernetes.Clientset, nss *NamespaceStorage) ResourceSource {
	source := ResourceSource{
		ItemType: "namespace",
		MapGet:   MapNamespaceGet,
		MapList:  MapNamespaceList,
	}

	source.LoadFunctions(
		cs.CoreV1().Namespaces().Get,
		cs.CoreV1().Namespaces().List,
	)

	return source
}

// MapNamespaceList maps an interface that is underneath a
// *coreV1.NamespaceList to a list of Items
func MapNamespaceList(i interface{}) ([]*sdp.Item, error) {
	var objectList *coreV1.NamespaceList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*coreV1.NamespaceList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *coreV1.NamespaceList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapNamespaceGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapNamespaceGet maps an interface that is underneath a *coreV1.Namespace to an item. If
// the interface isn't actually a *coreV1.Namespace this will fail
func MapNamespaceGet(i interface{}) (*sdp.Item, error) {
	var object *coreV1.Namespace
	var ok bool

	// Expect this to be a *coreV1.Namespace
	if object, ok = i.(*coreV1.Namespace); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *coreV1.Namespace", i)
	}

	item, err := mapK8sObject("namespace", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	context := strings.Join([]string{ClusterName, object.Name}, ".")

	// Link to all items in the namespace
	item.LinkedItemRequests = []*sdp.ItemRequest{
		// Search all types within the namespace's context
		{
			Context: context,
			Method:  sdp.RequestMethod_FIND,
			Type:    "*",
		},
	}

	return item, nil
}
