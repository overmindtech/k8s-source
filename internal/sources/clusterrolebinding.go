package sources

import (
	"fmt"
	"strings"

	rbacV1beta1 "k8s.io/api/rbac/v1beta1"

	"github.com/dylanratcliffe/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// ClusterRoleBindingSource returns a ResourceSource for ClusterRoleBindingClaims for a given
// client
func ClusterRoleBindingSource(cs *kubernetes.Clientset) ResourceSource {
	source := ResourceSource{
		ItemType:   "clusterrolebinding",
		MapGet:     MapClusterRoleBindingGet,
		MapList:    MapClusterRoleBindingList,
		Namespaced: false,
	}

	source.LoadFunction(
		cs.RbacV1beta1().ClusterRoleBindings,
	)

	return source
}

// MapClusterRoleBindingList maps an interface that is underneath a
// *rbacV1Beta1.ClusterRoleBindingList to a list of Items
func MapClusterRoleBindingList(i interface{}) ([]*sdp.Item, error) {
	var objectList *rbacV1beta1.ClusterRoleBindingList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*rbacV1beta1.ClusterRoleBindingList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *rbacV1Beta1.ClusterRoleBindingList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapClusterRoleBindingGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapClusterRoleBindingGet maps an interface that is underneath a *rbacV1Beta1.ClusterRoleBinding to an item. If
// the interface isn't actually a *rbacV1Beta1.ClusterRoleBinding this will fail
func MapClusterRoleBindingGet(i interface{}) (*sdp.Item, error) {
	var object *rbacV1beta1.ClusterRoleBinding
	var ok bool

	// Expect this to be a *rbacV1Beta1.ClusterRoleBinding
	if object, ok = i.(*rbacV1beta1.ClusterRoleBinding); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *rbacV1Beta1.ClusterRolebinding", i)
	}

	item, err := mapK8sObject("clusterrolebinding", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	var context string

	item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
		Context: ClusterName,
		Method:  sdp.RequestMethod_GET,
		Query:   object.RoleRef.Name,
		Type:    strings.ToLower(object.RoleRef.Kind),
	})

	for _, subject := range object.Subjects {
		if subject.Namespace == "" {
			context = ClusterName
		} else {
			context = ClusterName + "." + subject.Namespace
		}

		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Context: context,
			Method:  sdp.RequestMethod_GET,
			Query:   subject.Name,
			Type:    strings.ToLower(subject.Kind),
		})
	}

	return item, nil
}
