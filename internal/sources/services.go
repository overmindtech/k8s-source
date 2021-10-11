package sources

import (
	"fmt"

	"github.com/dylanratcliffe/sdp-go"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ServiceSource returns a ResourceSource for Pods for a given client and namespace
func ServiceSource(cs *kubernetes.Clientset) ResourceSource {
	source := ResourceSource{
		ItemType:   "service",
		MapGet:     MapServiceGet,
		MapList:    MapServiceList,
		Namespaced: true,
	}

	source.LoadFunction(
		cs.CoreV1().Services,
	)

	return source
}

// MapServiceList Maps the interface output of our list function to a list of
// items
func MapServiceList(i interface{}) ([]*sdp.Item, error) {
	var services *coreV1.ServiceList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	if services, ok = i.(*coreV1.ServiceList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *coreV1.ServiceList", i)
	}

	for _, service := range services.Items {
		if item, err = MapServiceGet(&service); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapServiceGet Maps an interface (which is the result of out Get function) to
// a service item
func MapServiceGet(i interface{}) (*sdp.Item, error) {
	var s *coreV1.Service
	var ok bool

	// Assert that i is a Service
	if s, ok = i.(*coreV1.Service); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *coreV1.Service", i)
	}

	item, err := mapK8sObject("service", s)

	if err != nil {
		return &sdp.Item{}, err
	}

	item.LinkedItemRequests = []*sdp.ItemRequest{
		{
			Context: item.Context,
			Method:  sdp.RequestMethod_GET,
			Query:   fmt.Sprint(item.UniqueAttributeValue()),
			Type:    "endpoint",
		},
	}

	if sel := s.Spec.Selector; sel != nil {
		// Services are linked to pods via their selector
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Context: item.Context,
			Method:  sdp.RequestMethod_SEARCH,
			Query: LabelSelectorToQuery(&metaV1.LabelSelector{
				MatchLabels: sel,
			}),
			Type: "pod",
		})
	}

	return item, nil
}
