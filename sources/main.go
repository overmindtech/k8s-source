package sources

import (
	"github.com/overmindtech/discovery"
	"k8s.io/client-go/kubernetes"
)

func LoadAllSources(cs *kubernetes.Clientset, cluster string, namespaces []string) []discovery.Source {
	return []discovery.Source{
		newClusterRoleSource(cs, cluster, namespaces),
		newClusterRoleBindingSource(cs, cluster, namespaces),
		newConfigMapSource(cs, cluster, namespaces),
		newCronJobSource(cs, cluster, namespaces),
		newDaemonSetSource(cs, cluster, namespaces),
		newDeploymentSource(cs, cluster, namespaces),
		newEndpointsSource(cs, cluster, namespaces),
		newEndpointSliceSource(cs, cluster, namespaces),
		newHorizontalPodAutoscalerSource(cs, cluster, namespaces),
		newIngressSource(cs, cluster, namespaces),
		newJobSource(cs, cluster, namespaces),
		newLimitRangeSource(cs, cluster, namespaces),
		newNetworkPolicySource(cs, cluster, namespaces),
		newNodeSource(cs, cluster, namespaces),
		newPersistentVolumeSource(cs, cluster, namespaces),
		newPersistentVolumeClaimSource(cs, cluster, namespaces),
		newPodDisruptionBudgetSource(cs, cluster, namespaces),
		newPodSource(cs, cluster, namespaces),
		newPriorityClassSource(cs, cluster, namespaces),
		newReplicaSetSource(cs, cluster, namespaces),
		newReplicationControllerSource(cs, cluster, namespaces),
		newResourceQuotaSource(cs, cluster, namespaces),
		newRoleSource(cs, cluster, namespaces),
		newRoleBindingSource(cs, cluster, namespaces),
		newSecretSource(cs, cluster, namespaces),
		newServiceSource(cs, cluster, namespaces),
		newServiceAccountSource(cs, cluster, namespaces),
		newStatefulSetSource(cs, cluster, namespaces),
		newStorageClassSource(cs, cluster, namespaces),
		newVolumeAttachmentSource(cs, cluster, namespaces),
	}
}
