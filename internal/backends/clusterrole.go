package backends

import (
	"fmt"

	rbacV1Beta1 "k8s.io/api/rbac/v1beta1"

	"github.com/dylanratcliffe/sdp-go"
	"k8s.io/client-go/kubernetes"
)

// ClusterRoleSource returns a ResourceSource for ClusterRoleClaims for a given
// client
func ClusterRoleSource(cs *kubernetes.Clientset, nss *NamespaceStorage) ResourceSource {
	source := ResourceSource{
		ItemType: "clusterrole",
		MapGet:   MapClusterRoleGet,
		MapList:  MapClusterRoleList,
	}

	source.LoadFunctions(
		cs.RbacV1beta1().ClusterRoles().Get,
		cs.RbacV1beta1().ClusterRoles().List,
	)

	return source
}

// MapClusterRoleList maps an interface that is underneath a
// *rbacV1Beta1.ClusterRoleList to a list of Items
func MapClusterRoleList(i interface{}) ([]*sdp.Item, error) {
	var objectList *rbacV1Beta1.ClusterRoleList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a objectList
	if objectList, ok = i.(*rbacV1Beta1.ClusterRoleList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *rbacV1Beta1.ClusterRoleList", i)
	}

	for _, object := range objectList.Items {
		if item, err = MapClusterRoleGet(&object); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapClusterRoleGet maps an interface that is underneath a *rbacV1Beta1.ClusterRole to an item. If
// the interface isn't actually a *rbacV1Beta1.ClusterRole this will fail
func MapClusterRoleGet(i interface{}) (*sdp.Item, error) {
	var object *rbacV1Beta1.ClusterRole
	var ok bool

	// Expect this to be a *rbacV1Beta1.ClusterRole
	if object, ok = i.(*rbacV1Beta1.ClusterRole); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *rbacV1Beta1.ClusterRole", i)
	}

	item, err := mapK8sObject("clusterrole", object)

	if err != nil {
		return &sdp.Item{}, err
	}

	return item, nil
}
