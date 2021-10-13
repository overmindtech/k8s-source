package sources

import (
	"fmt"
	"strings"

	"github.com/dylanratcliffe/sdp-go"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// PodSource returns a ResourceSource for Pods for a given client and namespace
func PodSource(cs *kubernetes.Clientset) (ResourceSource, error) {
	podsBackend := ResourceSource{
		ItemType:   "pod",
		MapGet:     MapPodGet,
		MapList:    MapPodList,
		Namespaced: true,
	}

	err := podsBackend.LoadFunction(
		cs.CoreV1().Pods,
	)

	return podsBackend, err
}

// MapPodList maps an interface that is underneath a *coreV1.PodList to a list of
// Items
func MapPodList(i interface{}) ([]*sdp.Item, error) {
	var podList *coreV1.PodList
	var ok bool
	var items []*sdp.Item
	var item *sdp.Item
	var err error

	// Expect this to be a podList
	if podList, ok = i.(*coreV1.PodList); !ok {
		return make([]*sdp.Item, 0), fmt.Errorf("could not convert %v to *coreV1.PodList", i)
	}

	for _, pod := range podList.Items {
		if item, err = MapPodGet(&pod); err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

// MapPodGet maps an interface that is underneath a *coreV1.Pod to an item. If the
// interface isn't actually a *coreV1.Pod this will fail
func MapPodGet(i interface{}) (*sdp.Item, error) {
	var pod *coreV1.Pod
	var ok bool

	// Expect this to be a pod
	if pod, ok = i.(*coreV1.Pod); !ok {
		return &sdp.Item{}, fmt.Errorf("could not assert %v as a *coreV1.Pod", i)
	}

	item, err := mapK8sObject("pod", pod)

	if err != nil {
		return &sdp.Item{}, err
	}

	// Link service accounts
	if pod.Spec.ServiceAccountName != "" {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Context: item.Context,
			Method:  sdp.RequestMethod_GET,
			Query:   pod.Spec.ServiceAccountName,
			Type:    "serviceaccount",
		})
	}

	// Link to the controller if relevant
	for _, ref := range pod.GetOwnerReferences() {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Context: item.Context,
			Type:    strings.ToLower(ref.Kind),
			Method:  sdp.RequestMethod_GET,
			Query:   ref.Name,
		})
	}

	// Link items from volumes
	for _, vol := range pod.Spec.Volumes {
		// Link PVCs
		if vol.PersistentVolumeClaim != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Context: item.Context,
				Method:  sdp.RequestMethod_GET,
				Query:   vol.PersistentVolumeClaim.ClaimName,
				Type:    PVCType,
			})
		}

		// Link secrets
		if vol.Secret != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Context: item.Context,
				Method:  sdp.RequestMethod_GET,
				Query:   vol.Secret.SecretName,
				Type:    "secret",
			})
		}

		// Link config map volumes
		if vol.ConfigMap != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Context: item.Context,
				Method:  sdp.RequestMethod_GET,
				Query:   vol.ConfigMap.Name,
				Type:    "configMap",
			})
		}
	}

	// Link items from containers
	for _, container := range pod.Spec.Containers {
		// Loop over environment variables
		for _, env := range container.Env {
			if env.ValueFrom != nil {
				if env.ValueFrom.SecretKeyRef != nil {
					// Add linked item from spec.containers[].env[].valueFrom.secretKeyRef
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Context: item.Context,
						Method:  sdp.RequestMethod_GET,
						Query:   env.ValueFrom.SecretKeyRef.Name,
						Type:    "secret",
					})
				}

				if env.ValueFrom.ConfigMapKeyRef != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Context: item.Context,
						Method:  sdp.RequestMethod_GET,
						Query:   env.ValueFrom.ConfigMapKeyRef.Name,
						Type:    "configMap",
					})
				}
			}
		}

		for _, envFrom := range container.EnvFrom {
			if envFrom.SecretRef != nil {
				// Add linked item from spec.containers[].EnvFrom[].secretKeyRef
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Context: item.Context,
					Method:  sdp.RequestMethod_GET,
					Query:   envFrom.SecretRef.Name,
					Type:    "secret",
				})
			}
		}
	}

	if pod.Spec.PriorityClassName != "" {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Context: ClusterName,
			Method:  sdp.RequestMethod_GET,
			Query:   pod.Spec.PriorityClassName,
			Type:    "priorityclassname",
		})
	}

	return item, nil
}
