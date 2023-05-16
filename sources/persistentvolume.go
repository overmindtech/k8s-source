package sources

import (
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func PersistentVolumeExtractor(resource *v1.PersistentVolume, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	sd, err := ParseScope(scope, false)

	if err != nil {
		return nil, err
	}

	if resource.Spec.PersistentVolumeSource.AWSElasticBlockStore != nil {
		// Link to EBS volume
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "ec2-volume",
				Method: sdp.QueryMethod_GET,
				Query:  resource.Spec.PersistentVolumeSource.AWSElasticBlockStore.VolumeID,
				Scope:  "*",
			},
		})
	}

	if resource.Spec.ClaimRef != nil {
		queries = append(queries, ObjectReferenceToQuery(resource.Spec.ClaimRef, sd))
	}

	if resource.Spec.StorageClassName != "" {
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "StorageClass",
				Method: sdp.QueryMethod_GET,
				Query:  resource.Spec.StorageClassName,
				Scope:  sd.ClusterName,
			},
		})
	}

	return queries, nil
}

func newPersistentVolumeSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.PersistentVolume, *v1.PersistentVolumeList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "PersistentVolume",
		ClusterInterfaceBuilder: func() ItemInterface[*v1.PersistentVolume, *v1.PersistentVolumeList] {
			return cs.CoreV1().PersistentVolumes()
		},
		ListExtractor: func(list *v1.PersistentVolumeList) ([]*v1.PersistentVolume, error) {
			extracted := make([]*v1.PersistentVolume, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: PersistentVolumeExtractor,
	}
}

func init() {
	registerSourceLoader(newPersistentVolumeSource)
}
