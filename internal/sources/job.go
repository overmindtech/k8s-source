package sources

import (
	"fmt"

	batchV1 "k8s.io/api/batch/v1"

	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// JobSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func JobSource(cs *kubernetes.Clientset) (ResourceSource, error) {
	source := ResourceSource{
		ItemType:   "job",
		MapGet:     MapJobGet,
		MapList:    MapJobList,
		Namespaced: true,
	}

	err := source.LoadFunction(
		cs.BatchV1().Jobs,
	)

	return source, err
}

// MapJobList maps an interface that is underneath a
// *batchV1.JobList to a list of Items
func MapJobList(i interface{}) ([]*sdp.Item, error) {
	var objectList *batchV1.JobList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*batchV1.JobList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *batchV1.JobList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapJobGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapJobGet maps an interface that is underneath a *batchV1.Job to an item. If
// the interface isn't actually a *batchV1.Job this will fail
func MapJobGet(i interface{}) (*sdp.Item, error) {
	var object *batchV1.Job
	var ok bool

	// Expect this to be a *batchV1.Job
	if object, ok = i.(*batchV1.Job); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *batchV1.Job", i)
	}

	item, err := mapK8sObject("job", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	if object.Spec.Selector != nil {
		item.LinkedItemRequests = []*sdp.ItemRequest{
			{
				Context: item.Context,
				Method:  sdp.RequestMethod_SEARCH,
				Query:   LabelSelectorToQuery(object.Spec.Selector),
				Type:    "pod",
			},
		}
	}

	// Check owner references to see if it was created by a cronjob
	for _, o := range object.ObjectMeta.OwnerReferences {
		if o.Kind == "CronJob" {
			item.LinkedItemRequests = []*sdp.ItemRequest{
				{
					Context: item.Context,
					Method:  sdp.RequestMethod_GET,
					Query:   o.Name,
					Type:    "cronjob",
				},
			}
		}
	}

	return item, nil
}
