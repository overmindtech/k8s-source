package backends

import (
	"fmt"
	"strings"

	"github.com/dylanratcliffe/sdp-go"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

// NodeSource returns a ResourceSource for PersistentVolumeClaims for a given
// client
func NodeSource(cs *kubernetes.Clientset, nss *NamespaceStorage) ResourceSource {
	source := ResourceSource{
		ItemType: "node",
		NSS:      nss,
	}

	source.MapGet = source.MapNodeGet
	source.MapList = source.MapNodeList

	source.LoadFunctions(
		cs.CoreV1().Nodes().Get,
		cs.CoreV1().Nodes().List,
	)

	return source
}

// MapNodeList maps an interface that is underneath a
// *coreV1.NodeList to a list of Items
func (rs *ResourceSource) MapNodeList(i interface{}) ([]*sdp.Item, error) {
	var objectList *coreV1.NodeList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*coreV1.NodeList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *coreV1.NodeList", i)
	}

	for _, object := range objectList.Items {
		if item, err = rs.MapNodeGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapNodeGet maps an interface that is underneath a *coreV1.Node to an item. If
// the interface isn't actually a *coreV1.Node this will fail
func (rs *ResourceSource) MapNodeGet(i interface{}) (*sdp.Item, error) {
	var object *coreV1.Node
	var ok bool

	// Expect this to be a *coreV1.Node
	if object, ok = i.(*coreV1.Node); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *coreV1.Node", i)
	}

	item, err := mapK8sObject("node", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	// Query based onf fields not labels
	hostQuery := metaV1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%v", object.Name),
	}

	namespaces, _ := rs.NSS.Namespaces()

	for _, namespace := range namespaces {
		context := strings.Join([]string{ClusterName, namespace}, ".")

		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Context: context,
			Method:  sdp.RequestMethod_SEARCH,
			Type:    "pod",
			Query:   ListOptionsToQuery(&hostQuery),
		})
	}

	return item, nil
}
