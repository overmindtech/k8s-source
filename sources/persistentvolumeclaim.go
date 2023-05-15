package sources

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func newPersistentVolumeClaimSource(cs *kubernetes.Clientset, cluster string, namespaces []string) *KubeTypeSource[*v1.PersistentVolumeClaim, *v1.PersistentVolumeClaimList] {
	return &KubeTypeSource[*v1.PersistentVolumeClaim, *v1.PersistentVolumeClaimList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "PersistentVolumeClaim",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.PersistentVolumeClaim, *v1.PersistentVolumeClaimList] {
			return cs.CoreV1().PersistentVolumeClaims(namespace)
		},
		ListExtractor: func(list *v1.PersistentVolumeClaimList) ([]*v1.PersistentVolumeClaim, error) {
			extracted := make([]*v1.PersistentVolumeClaim, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
	}
}
