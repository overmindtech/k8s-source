package sources

import (
	"fmt"
	"strings"

	autoscalingV1 "k8s.io/api/autoscaling/v1"

	"github.com/dylanratcliffe/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// HorizontalPodAutoscalerSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func HorizontalPodAutoscalerSource(cs *kubernetes.Clientset) (ResourceSource, error) {
	source := ResourceSource{
		ItemType:   "horizontalpodautoscaler",
		MapGet:     MapHorizontalPodAutoscalerGet,
		MapList:    MapHorizontalPodAutoscalerList,
		Namespaced: true,
	}

	err := source.LoadFunction(
		cs.AutoscalingV1().HorizontalPodAutoscalers,
	)

	return source, err
}

// MapHorizontalPodAutoscalerList maps an interface that is underneath a
// *autoscalingV1.HorizontalPodAutoscalerList to a list of Items
func MapHorizontalPodAutoscalerList(i interface{}) ([]*sdp.Item, error) {
	var objectList *autoscalingV1.HorizontalPodAutoscalerList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*autoscalingV1.HorizontalPodAutoscalerList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *autoscalingV1.HorizontalPodAutoscalerList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapHorizontalPodAutoscalerGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapHorizontalPodAutoscalerGet maps an interface that is underneath a *autoscalingV1.HorizontalPodAutoscaler to an item. If
// the interface isn't actually a *autoscalingV1.HorizontalPodAutoscaler this will fail
func MapHorizontalPodAutoscalerGet(i interface{}) (*sdp.Item, error) {
	var object *autoscalingV1.HorizontalPodAutoscaler
	var ok bool

	// Expect this to be a *autoscalingV1.HorizontalPodAutoscaler
	if object, ok = i.(*autoscalingV1.HorizontalPodAutoscaler); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *autoscalingV1.HorizontalPodAutoscaler", i)
	}

	item, err := mapK8sObject("horizontalpodautoscaler", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	item.LinkedItemRequests = []*sdp.ItemRequest{
		// Services are linked to pods via their selector
		{
			Context: item.Context,
			Method:  sdp.RequestMethod_GET,
			Query:   object.Spec.ScaleTargetRef.Name,
			Type:    strings.ToLower(object.Spec.ScaleTargetRef.Kind),
		},
	}

	return item, nil
}
