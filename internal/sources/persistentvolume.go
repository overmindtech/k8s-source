package sources

import (
	"fmt"
	"strings"

	"github.com/overmindtech/sdp-go"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// PersistentVolumeSource returns a ResourceSource for PersistentVolumeClaims for a given
// client
func PersistentVolumeSource(cs *kubernetes.Clientset) (ResourceSource, error) {
	source := ResourceSource{
		ItemType:   "persistentvolume",
		MapGet:     MapPersistentVolumeGet,
		MapList:    MapPersistentVolumeList,
		Namespaced: false,
	}

	err := source.LoadFunction(
		cs.CoreV1().PersistentVolumes,
	)

	return source, err
}

// MapPersistentVolumeList maps an interface that is underneath a
// *coreV1.PersistentVolumeList to a list of Items
func MapPersistentVolumeList(i interface{}) ([]*sdp.Item, error) {
	var objectList *coreV1.PersistentVolumeList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*coreV1.PersistentVolumeList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *coreV1.PersistentVolumeList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapPersistentVolumeGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapPersistentVolumeGet maps an interface that is underneath a *coreV1.PersistentVolume to an item. If
// the interface isn't actually a *coreV1.PersistentVolume this will fail
func MapPersistentVolumeGet(i interface{}) (*sdp.Item, error) {
	var object *coreV1.PersistentVolume
	var ok bool

	// Expect this to be a *coreV1.PersistentVolume
	if object, ok = i.(*coreV1.PersistentVolume); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *coreV1.PersistentVolume", i)
	}

	item, err := mapK8sObject("persistentvolume", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	if claim := object.Spec.ClaimRef; claim != nil {
		context := strings.Join([]string{ClusterName, claim.Namespace}, ".")

		// Link to all items in the PersistentVolume
		item.LinkedItemRequests = []*sdp.ItemRequest{
			// Search all types within the PersistentVolume's context
			{
				Context: context,
				Method:  sdp.RequestMethod_GET,
				Type:    strings.ToLower(claim.Kind),
				Query:   claim.Name,
			},
		}
	}

	// Link to the storage class
	item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
		Context: ClusterName,
		Method:  sdp.RequestMethod_GET,
		Query:   object.Spec.StorageClassName,
		Type:    "storageclass",
	})

	return item, nil
}
