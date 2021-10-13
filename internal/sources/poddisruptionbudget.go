package sources

import (
	"fmt"

	"github.com/dylanratcliffe/sdp-go"
	policyV1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/client-go/kubernetes"
)

// PodDisruptionBudgetSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func PodDisruptionBudgetSource(cs *kubernetes.Clientset) (ResourceSource, error) {
	source := ResourceSource{
		ItemType:   "poddisruptionbudget",
		MapGet:     MapPodDisruptionBudgetGet,
		MapList:    MapPodDisruptionBudgetList,
		Namespaced: true,
	}

	err := source.LoadFunction(
		cs.PolicyV1beta1().PodDisruptionBudgets,
	)

	return source, err
}

// MapPodDisruptionBudgetList maps an interface that is underneath a
// *policyV1beta1.PodDisruptionBudgetList to a list of Items
func MapPodDisruptionBudgetList(i interface{}) ([]*sdp.Item, error) {
	var objectList *policyV1beta1.PodDisruptionBudgetList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*policyV1beta1.PodDisruptionBudgetList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *policyV1beta1.PodDisruptionBudgetList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapPodDisruptionBudgetGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapPodDisruptionBudgetGet maps an interface that is underneath a *policyV1beta1.PodDisruptionBudget to an item. If
// the interface isn't actually a *policyV1beta1.PodDisruptionBudget this will fail
func MapPodDisruptionBudgetGet(i interface{}) (*sdp.Item, error) {
	var object *policyV1beta1.PodDisruptionBudget
	var ok bool

	// Expect this to be a *policyV1beta1.PodDisruptionBudget
	if object, ok = i.(*policyV1beta1.PodDisruptionBudget); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *policyV1beta1.PodDisruptionBudget", i)
	}

	item, err := mapK8sObject("poddisruptionbudget", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	if selector := object.Spec.Selector; selector != nil {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Context: item.Context,
			Method:  sdp.RequestMethod_SEARCH,
			Query:   LabelSelectorToQuery(selector),
			Type:    "pod",
		})
	}

	return item, nil
}
