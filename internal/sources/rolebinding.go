package sources

import (
	"fmt"
	"strings"

	rbacV1 "k8s.io/api/rbac/v1"

	"github.com/overmindtech/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// RoleBindingSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func RoleBindingSource(cs *kubernetes.Clientset) (ResourceSource, error) {
	source := ResourceSource{
		ItemType:   "rolebinding",
		MapGet:     MapRoleBindingGet,
		MapList:    MapRoleBindingList,
		Namespaced: true,
	}

	err := source.LoadFunction(
		cs.RbacV1().RoleBindings,
	)

	return source, err
}

// MapRoleBindingList maps an interface that is underneath a
// *rbacV1.RoleBindingList to a list of Items
func MapRoleBindingList(i interface{}) ([]*sdp.Item, error) {
	var objectList *rbacV1.RoleBindingList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*rbacV1.RoleBindingList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *rbacV1.RoleBindingList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapRoleBindingGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapRoleBindingGet maps an interface that is underneath a *rbacV1.RoleBinding to an item. If
// the interface isn't actually a *rbacV1.RoleBinding this will fail
func MapRoleBindingGet(i interface{}) (*sdp.Item, error) {
	var object *rbacV1.RoleBinding
	var ok bool

	// Expect this to be a *rbacV1.RoleBinding
	if object, ok = i.(*rbacV1.RoleBinding); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *rbacV1.RoleBinding", i)
	}

	item, err := mapK8sObject("rolebinding", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	// Link the referenced role
	var scope string

	switch object.RoleRef.Name {
	case "Role":
		// If this binding is linked to a role then it's in the same namespace
		scope = item.Scope
	case "ClusterRole":
		// If this is linked to a ClusterRole (which is not namespaced) we need
		// to make sure that we are querying the root scope i.e. the
		// non-namespaced scope
		scope = ClusterName
	}

	item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
		Scope:  scope,
		Method: sdp.QueryMethod_GET,
		Query:  object.RoleRef.Name,
		Type:   strings.ToLower(object.RoleRef.Kind),
	})

	for _, subject := range object.Subjects {
		if subject.Namespace == "" {
			scope = ClusterName
		} else {
			scope = ClusterName + "." + subject.Namespace
		}

		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
			Scope:  scope,
			Method: sdp.QueryMethod_GET,
			Query:  subject.Name,
			Type:   strings.ToLower(subject.Kind),
		})
	}

	return item, nil
}
