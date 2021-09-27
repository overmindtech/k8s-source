package backends

import (
	"fmt"

	"github.com/dylanratcliffe/sdp-go"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// PVCType is the name of the PVC type. I'm saving this as a const since it's a
// bit nasty and I might want to change it later
const PVCType = "persistentVolumeClaim"

// PVCSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func PVCSource(cs *kubernetes.Clientset, namespace string) ResourceSource {
	source := ResourceSource{
		ItemType: PVCType,
		MapGet:   MapPVCGet,
		MapList:  MapPVCList,
	}

	source.LoadFunctions(
		cs.CoreV1().PersistentVolumeClaims(namespace).Get,
		cs.CoreV1().PersistentVolumeClaims(namespace).List,
	)

	return source
}

// MapPVCList maps an interface that is underneath a
// *coreV1.PersistentVolumeClaimList to a list of Items
func MapPVCList(i interface{}) ([]*sdp.Item, error) {
	var pvcList *coreV1.PersistentVolumeClaimList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a pvcList
	if pvcList, ok = i.(*coreV1.PersistentVolumeClaimList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *coreV1.PersistentVolumeClaimList", i)
	}

	for _, pvc := range pvcList.Items {
		if item, err = MapPVCGet(&pvc); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapPVCGet maps an interface that is underneath a *coreV1.PersistentVolumeClaim to
// an item. If the interface isn't actually a *coreV1.PersistentVolumeClaim this
// will fail
func MapPVCGet(i interface{}) (*sdp.Item, error) {
	var pvc *coreV1.PersistentVolumeClaim
	var ok bool

	// Expect this to be a pvc
	if pvc, ok = i.(*coreV1.PersistentVolumeClaim); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *coreV1.PersistentVolumeClaim", i)
	}

	item, err := mapK8sObject(PVCType, pvc)

	if err != nil {
		return &sdp.Item{}, err
	}

	return item, nil
}
