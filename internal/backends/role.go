package backends

import (
	"fmt"

	rbacV1 "k8s.io/api/rbac/v1"

	"github.com/dylanratcliffe/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// RoleSource returns a ResourceSource for PersistentVolumeClaims for a given
// client and namespace
func RoleSource(cs *kubernetes.Clientset, namespace string) ResourceSource {
	source := ResourceSource{
		ItemType: "role",
		MapGet:   MapRoleGet,
		MapList:  MapRoleList,
	}

	source.LoadFunctions(
		cs.RbacV1().Roles(namespace).Get,
		cs.RbacV1().Roles(namespace).List,
	)

	return source
}

// MapRoleList maps an interface that is underneath a
// *rbacV1.RoleList to a list of Items
func MapRoleList(i interface{}) ([]*sdp.Item, error) {
	var objectList *rbacV1.RoleList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*rbacV1.RoleList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *rbacV1.RoleList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapRoleGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapRoleGet maps an interface that is underneath a *rbacV1.Role to an item. If
// the interface isn't actually a *rbacV1.Role this will fail
func MapRoleGet(i interface{}) (*sdp.Item, error) {
	var object *rbacV1.Role
	var ok bool

	// Expect this to be a *rbacV1.Role
	if object, ok = i.(*rbacV1.Role); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *rbacV1.Role", i)
	}

	item, err := mapK8sObject("role", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	return item, nil
}
