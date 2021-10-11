package sources

import (
	"fmt"

	batchV1beta1 "k8s.io/api/batch/v1beta1"

	"github.com/dylanratcliffe/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// CronJobSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func CronJobSource(cs *kubernetes.Clientset, namespace string) ResourceSource {
	source := ResourceSource{
		ItemType: "cronjob",
		MapGet:   MapCronJobGet,
		MapList:  MapCronJobList,
	}

	source.LoadFunctions(
		cs.BatchV1beta1().CronJobs(namespace).Get,
		cs.BatchV1beta1().CronJobs(namespace).List,
	)

	return source
}

// MapCronJobList maps an interface that is underneath a
// *batchV1beta1.CronJobList to a list of Items
func MapCronJobList(i interface{}) ([]*sdp.Item, error) {
	var objectList *batchV1beta1.CronJobList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*batchV1beta1.CronJobList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *batchV1beta1.CronJobList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapCronJobGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapCronJobGet maps an interface that is underneath a *batchV1beta1.CronJob to an item. If
// the interface isn't actually a *batchV1beta1.CronJob this will fail
func MapCronJobGet(i interface{}) (*sdp.Item, error) {
	var object *batchV1beta1.CronJob
	var ok bool

	// Expect this to be a *batchV1beta1.CronJob
	if object, ok = i.(*batchV1beta1.CronJob); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *batchV1beta1.CronJob", i)
	}

	item, err := mapK8sObject("cronjob", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	return item, nil
}
