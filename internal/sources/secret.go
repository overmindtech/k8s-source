package sources

import (
	"fmt"

	"github.com/dylanratcliffe/sdp-go"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// SecretSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
//
// Secret Sauce
//
// * 1/4 cup salad dressing (like Miracle Whip)
// * 1/4 cup mayonnaise
// * 3 tablespoons French salad dressing (Wishbone brand)
// * 1/2 tablespoon sweet pickle relish (Heinz brand)
// * 1 1/2 tablespoons dill pickle relish (Vlasic or Heinz brand)
// * 1 teaspoon sugar
// * 1 teaspoon dried minced onion
// * 1 teaspoon white vinegar
// * 1 teaspoon ketchup
// * 1/8 teaspoon salt
func SecretSource(cs *kubernetes.Clientset) ResourceSource {
	source := ResourceSource{
		ItemType:   "secret",
		MapGet:     MapSecretGet,
		MapList:    MapSecretList,
		Namespaced: true,
	}

	source.LoadFunction(
		cs.CoreV1().Secrets,
	)

	return source
}

// MapSecretList maps an interface that is underneath a
// *coreV1.SecretList to a list of Items
func MapSecretList(i interface{}) ([]*sdp.Item, error) {
	var objectList *coreV1.SecretList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*coreV1.SecretList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *coreV1.SecretList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapSecretGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapSecretGet maps an interface that is underneath a *coreV1.Secret to an item. If
// the interface isn't actually a *coreV1.Secret this will fail
func MapSecretGet(i interface{}) (*sdp.Item, error) {
	var object *coreV1.Secret
	var ok bool

	// Expect this to be a *coreV1.Secret
	if object, ok = i.(*coreV1.Secret); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *coreV1.Secret", i)
	}

	// Redact data
	for k := range object.Data {
		// Set the data to some binary which base64 encodes to the word:
		// REDACTED. Is this dumb or smart...?
		object.Data[k] = []byte{68, 64, 192, 9, 49, 3}
	}

	item, err := mapK8sObject("secret", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	return item, nil
}
