package adapters

import (
	"regexp"

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
		// +overmind:link ec2-volume
		// Link to EBS volume
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "ec2-volume",
				Method: sdp.QueryMethod_GET,
				Query:  resource.Spec.PersistentVolumeSource.AWSElasticBlockStore.VolumeID,
				Scope:  "*",
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changes to the EBS volume can affect the PV
				In: true,
				// Changes to the PV might affect the EBS volume
				Out: true,
			},
		})
	}

	if resource.Spec.CSI != nil {
		// Link to an EFS file system access point
		efsVolumeHandle := regexp.MustCompile(`fs-[a-f0-9]+::(fsap-[a-f0-9]+)`)

		matches := efsVolumeHandle.FindStringSubmatch(resource.Spec.CSI.VolumeHandle)

		if matches != nil {
			if len(matches) == 2 {
				// +overmind:link efs-access-point
				queries = append(queries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "efs-access-point",
						Method: sdp.QueryMethod_GET,
						Query:  matches[1],
						Scope:  "*",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to the EFS access point can affect the PV
						In: true,
						// Changes to the PV won't affect the EFS access point
						Out: false,
					},
				})
			}
		}
	}

	if resource.Spec.ClaimRef != nil {
		queries = append(queries, ObjectReferenceToQuery(resource.Spec.ClaimRef, sd, &sdp.BlastPropagation{
			// Changing claim might not affect the PV
			In: false,
			// Changing the PV will definitely affect the claim
			Out: true,
		}))
	}

	if resource.Spec.StorageClassName != "" {
		// +overmind:link StorageClass
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "StorageClass",
				Method: sdp.QueryMethod_GET,
				Query:  resource.Spec.StorageClassName,
				Scope:  sd.ClusterName,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changes to the storage class can affect the PV
				In: true,
				// Changes to the PV cannot affect the storage class
				Out: false,
			},
		})
	}

	return queries, nil
}

//go:generate docgen ../docs-data
// +overmind:type PersistentVolume
// +overmind:descriptiveType Persistent Volume
// +overmind:get Get a persistent volume by name
// +overmind:list List all persistent volumes
// +overmind:search Search for a persistent volume using the ListOptions JSON format: https://github.com/overmindtech/k8s-source#search
// +overmind:group Kubernetes
// +overmind:terraform:queryMap kubernetes_persistent_volume.metadata[0].name
// +overmind:terraform:queryMap kubernetes_persistent_volume_v1.metadata[0].name
// +overmind:terraform:scope ${provider_mapping.cluster_name}.${values.metadata[0].namespace}

func newPersistentVolumeAdapter(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Adapter {
	return &KubeTypeAdapter[*v1.PersistentVolume, *v1.PersistentVolumeList]{
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
	registerAdapterLoader(newPersistentVolumeAdapter)
}
