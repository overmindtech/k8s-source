package sources

import (
	"fmt"

	"github.com/overmindtech/sdp-go"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// ServiceAccountSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func ServiceAccountSource(cs *kubernetes.Clientset) (ResourceSource, error) {
	source := ResourceSource{
		ItemType:   "serviceaccount",
		MapGet:     MapServiceAccountGet,
		MapList:    MapServiceAccountList,
		Namespaced: true,
	}

	err := source.LoadFunction(
		cs.CoreV1().ServiceAccounts,
	)

	return source, err
}

// MapServiceAccountList maps an interface that is underneath a
// *coreV1.ServiceAccountList to a list of Items
func MapServiceAccountList(i interface{}) ([]*sdp.Item, error) {
	var objectList *coreV1.ServiceAccountList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*coreV1.ServiceAccountList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *coreV1.ServiceAccountList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapServiceAccountGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapServiceAccountGet maps an interface that is underneath a *coreV1.ServiceAccount to an item. If
// the interface isn't actually a *coreV1.ServiceAccount this will fail
func MapServiceAccountGet(i interface{}) (*sdp.Item, error) {
	var object *coreV1.ServiceAccount
	var ok bool

	// Expect this to be a *coreV1.ServiceAccount
	if object, ok = i.(*coreV1.ServiceAccount); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *coreV1.ServiceAccount", i)
	}

	item, err := mapK8sObject("serviceaccount", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	for _, secret := range object.Secrets {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Context: item.Context,
			Method:  sdp.RequestMethod_GET,
			Query:   secret.Name,
			Type:    "secret",
		})
	}

	for _, ipSecret := range object.ImagePullSecrets {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Context: item.Context,
			Method:  sdp.RequestMethod_GET,
			Query:   ipSecret.Name,
			Type:    "secret",
		})
	}

	return item, nil
}
