package sources

import (
	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func PodExtractor(resource *v1.Pod, scope string) ([]*sdp.Query, error) {
	queries := make([]*sdp.Query, 0)

	// Link service accounts
	if resource.Spec.ServiceAccountName != "" {
		queries = append(queries, &sdp.Query{
			Scope:  scope,
			Method: sdp.QueryMethod_GET,
			Query:  resource.Spec.ServiceAccountName,
			Type:   "ServiceAccount",
		})
	}

	// Link items from volumes
	for _, vol := range resource.Spec.Volumes {
		// Link PVCs
		if vol.PersistentVolumeClaim != nil {
			queries = append(queries, &sdp.Query{
				Scope:  scope,
				Method: sdp.QueryMethod_GET,
				Query:  vol.PersistentVolumeClaim.ClaimName,
				Type:   "PersistentVolumeClaim",
			})
		}

		// Link secrets
		if vol.Secret != nil {
			queries = append(queries, &sdp.Query{
				Scope:  scope,
				Method: sdp.QueryMethod_GET,
				Query:  vol.Secret.SecretName,
				Type:   "Secret",
			})
		}

		// Link config map volumes
		if vol.ConfigMap != nil {
			queries = append(queries, &sdp.Query{
				Scope:  scope,
				Method: sdp.QueryMethod_GET,
				Query:  vol.ConfigMap.Name,
				Type:   "ConfigMap",
			})
		}
	}

	// Link items from containers
	for _, container := range resource.Spec.Containers {
		// Loop over environment variables
		for _, env := range container.Env {
			if env.ValueFrom != nil {
				if env.ValueFrom.SecretKeyRef != nil {
					// Add linked item from spec.containers[].env[].valueFrom.secretKeyRef
					queries = append(queries, &sdp.Query{
						Scope:  scope,
						Method: sdp.QueryMethod_GET,
						Query:  env.ValueFrom.SecretKeyRef.Name,
						Type:   "Secret",
					})
				}

				if env.ValueFrom.ConfigMapKeyRef != nil {
					queries = append(queries, &sdp.Query{
						Scope:  scope,
						Method: sdp.QueryMethod_GET,
						Query:  env.ValueFrom.ConfigMapKeyRef.Name,
						Type:   "ConfigMap",
					})
				}
			}
		}

		for _, envFrom := range container.EnvFrom {
			if envFrom.SecretRef != nil {
				// Add linked item from spec.containers[].EnvFrom[].secretKeyRef
				queries = append(queries, &sdp.Query{
					Scope:  scope,
					Method: sdp.QueryMethod_GET,
					Query:  envFrom.SecretRef.Name,
					Type:   "Secret",
				})
			}
		}
	}

	if resource.Spec.PriorityClassName != "" {
		queries = append(queries, &sdp.Query{
			Scope:  ClusterName,
			Method: sdp.QueryMethod_GET,
			Query:  resource.Spec.PriorityClassName,
			Type:   "PriorityClass",
		})
	}

	if len(resource.Status.PodIPs) > 0 {
		for _, ip := range resource.Status.PodIPs {
			queries = append(queries, &sdp.Query{
				Scope:  "global",
				Method: sdp.QueryMethod_GET,
				Query:  ip.IP,
				Type:   "ip",
			})
		}
	} else if resource.Status.PodIP != "" {
		queries = append(queries, &sdp.Query{
			Type:   "ip",
			Method: sdp.QueryMethod_GET,
			Query:  resource.Status.PodIP,
			Scope:  "global",
		})
	}

	return queries, nil
}

func NewPodSource(cs *kubernetes.Clientset, cluster string, namespaces []string) KubeTypeSource[*v1.Pod, *v1.PodList] {
	return KubeTypeSource[*v1.Pod, *v1.PodList]{
		ClusterName: cluster,
		Namespaces:  namespaces,
		TypeName:    "Pod",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.Pod, *v1.PodList] {
			return cs.CoreV1().Pods(namespace)
		},
		ListExtractor: func(list *v1.PodList) ([]*v1.Pod, error) {
			extracted := make([]*v1.Pod, len(list.Items))

			for i := range list.Items {
				extracted[i] = &list.Items[i]
			}

			return extracted, nil
		},
		LinkedItemQueryExtractor: PodExtractor,
		HealthExtractor: func(resource *v1.Pod) *sdp.Health {
			switch resource.Status.Phase {
			case v1.PodPending:
				return sdp.Health_HEALTH_PENDING.Enum()
			case v1.PodRunning:
				return sdp.Health_HEALTH_OK.Enum()
			case v1.PodSucceeded:
				return sdp.Health_HEALTH_OK.Enum()
			case v1.PodFailed:
				return sdp.Health_HEALTH_ERROR.Enum()
			case v1.PodUnknown:
				return sdp.Health_HEALTH_UNKNOWN.Enum()
			}

			return nil
		},
	}
}
