package sources

import (
	"fmt"

	appsV1 "k8s.io/api/apps/v1"

	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// DeploymentSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func DeploymentSource(cs *kubernetes.Clientset) (ResourceSource, error) {
	source := ResourceSource{
		ItemType:   "deployment",
		MapGet:     MapDeploymentGet,
		MapList:    MapDeploymentList,
		Namespaced: true,
	}

	err := source.LoadFunction(
		cs.AppsV1().Deployments,
	)

	return source, err
}

// MapDeploymentList maps an interface that is underneath a
// *appsV1.DeploymentList to a list of Items
func MapDeploymentList(i interface{}) ([]*sdp.Item, error) {
	var objectList *appsV1.DeploymentList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*appsV1.DeploymentList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *appsV1.DeploymentList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapDeploymentGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapDeploymentGet maps an interface that is underneath a *appsV1.Deployment to an item. If
// the interface isn't actually a *appsV1.Deployment this will fail
func MapDeploymentGet(i interface{}) (*sdp.Item, error) {
	var object *appsV1.Deployment
	var ok bool

	// Expect this to be a *appsV1.Deployment
	if object, ok = i.(*appsV1.Deployment); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *appsV1.Deployment", i)
	}

	item, err := mapK8sObject("deployment", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	if object.Spec.Selector != nil {
		item.LinkedItemQueries = []*sdp.Query{
			// Services are linked to pods via their selector
			{
				Scope:  item.Scope,
				Method: sdp.QueryMethod_SEARCH,
				Query:  LabelSelectorToQuery(object.Spec.Selector),
				Type:   "replicaset",
			},
		}
	}

	return item, nil
}
