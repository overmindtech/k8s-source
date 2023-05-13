package sources

import (
	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/storage/v1"
	"k8s.io/client-go/kubernetes"
)

func volumeAttachmentExtractor(resource *v1.VolumeAttachment, scope string) ([]*sdp.Query, error) {
	queries := make([]*sdp.Query, 0)

	if resource.Spec.Source.PersistentVolumeName != nil {
		queries = append(queries, &sdp.Query{
			Type:   "PersistentVolume",
			Method: sdp.QueryMethod_GET,
			Query:  *resource.Spec.Source.PersistentVolumeName,
			Scope:  scope,
		})
	}

	if resource.Spec.NodeName != "" {
		queries = append(queries, &sdp.Query{
			Type:   "Node",
			Method: sdp.QueryMethod_GET,
			Query:  resource.Spec.NodeName,
			Scope:  scope,
		})
	}

	return queries, nil
}

func NewVolumeAttachmentSource(cs *kubernetes.Clientset, cluster string, namespaces []string) KubeTypeSource[*v1.VolumeAttachment, *v1.VolumeAttachmentList] {
	return KubeTypeSource[*v1.VolumeAttachment, *v1.VolumeAttachmentList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "VolumeAttachment",
		ClusterInterfaceBuilder: func() ItemInterface[*v1.VolumeAttachment, *v1.VolumeAttachmentList] {
			return cs.StorageV1().VolumeAttachments()
		},
		ListExtractor: func(list *v1.VolumeAttachmentList) ([]*v1.VolumeAttachment, error) {
			extracted := make([]*v1.VolumeAttachment, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: volumeAttachmentExtractor,
		HealthExtractor: func(resource *v1.VolumeAttachment) *sdp.Health {
			if resource.Status.AttachError != nil || resource.Status.DetachError != nil {
				return sdp.Health_HEALTH_ERROR.Enum()
			}

			return sdp.Health_HEALTH_OK.Enum()
		},
	}
}
