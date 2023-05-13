package sources

import (
	v1 "k8s.io/api/storage/v1"

	"k8s.io/client-go/kubernetes"
)

func NewStorageClassSource(cs *kubernetes.Clientset, cluster string, namespaces []string) KubeTypeSource[*v1.StorageClass, *v1.StorageClassList] {
	return KubeTypeSource[*v1.StorageClass, *v1.StorageClassList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "StorageClass",
		ClusterInterfaceBuilder: func() ItemInterface[*v1.StorageClass, *v1.StorageClassList] {
			return cs.StorageV1().StorageClasses()
		},
		ListExtractor: func(list *v1.StorageClassList) ([]*v1.StorageClass, error) {
			extracted := make([]*v1.StorageClass, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
	}
}
