package sources

import (
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func PodExtractor(resource *v1.Pod, scope string) ([]*sdp.LinkedItemQuery, error) {
	queries := make([]*sdp.LinkedItemQuery, 0)

	sd, err := ParseScope(scope, true)

	if err != nil {
		return nil, err
	}

	// Link service accounts
	if resource.Spec.ServiceAccountName != "" {
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "ServiceAccount",
				Scope:  scope,
				Method: sdp.QueryMethod_GET,
				Query:  resource.Spec.ServiceAccountName,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changes to the service account can affect the pod
				In: true,
				// Changes to the pod cannot affect the service account
				Out: false,
			},
		})
	}

	// Link items from volumes
	for _, vol := range resource.Spec.Volumes {
		// Link PVCs
		if vol.PersistentVolumeClaim != nil {
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Scope:  scope,
					Method: sdp.QueryMethod_GET,
					Query:  vol.PersistentVolumeClaim.ClaimName,
					Type:   "PersistentVolumeClaim",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the PVC will affect the pod
					In: true,
					// The pod can definitely affect the PVC, by filling it up
					// for example
					Out: true,
				},
			})
		}

		// Link secrets
		if vol.Secret != nil {
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Scope:  scope,
					Method: sdp.QueryMethod_GET,
					Query:  vol.Secret.SecretName,
					Type:   "Secret",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the secret could easily break the pod
					In: true,
					// The pod however isn't going to affect a secret
					Out: false,
				},
			})
		}

		// Link config map volumes
		if vol.ConfigMap != nil {
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Scope:  scope,
					Method: sdp.QueryMethod_GET,
					Query:  vol.ConfigMap.Name,
					Type:   "ConfigMap",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the config map could easily break the pod
					In: true,
					// The pod however isn't going to affect a config map
					Out: false,
				},
			})
		}

		// Link projected volumes
		if vol.Projected != nil {
			for _, source := range vol.Projected.Sources {
				if source.ConfigMap != nil {
					queries = append(queries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Scope:  scope,
							Method: sdp.QueryMethod_GET,
							Query:  source.ConfigMap.Name,
							Type:   "ConfigMap",
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the config map could easily break the pod
							In: true,
							// The pod however isn't going to affect a config map
							Out: false,
						},
					})
				}

				if source.Secret != nil {
					queries = append(queries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Scope:  scope,
							Method: sdp.QueryMethod_GET,
							Query:  source.Secret.Name,
							Type:   "Secret",
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the secret could easily break the pod
							In: true,
							// The pod however isn't going to affect a secret
							Out: false,
						},
					})
				}
			}
		}
	}

	// Link items from containers
	for _, container := range resource.Spec.Containers {
		// Loop over environment variables
		for _, env := range container.Env {
			if env.ValueFrom != nil {
				if env.ValueFrom.SecretKeyRef != nil {
					// Add linked item from spec.containers[].env[].valueFrom.secretKeyRef
					queries = append(queries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Scope:  scope,
							Method: sdp.QueryMethod_GET,
							Query:  env.ValueFrom.SecretKeyRef.Name,
							Type:   "Secret",
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the secret could easily break the pod
							In: true,
							// The pod however isn't going to affect a secret
							Out: false,
						},
					})
				}

				if env.ValueFrom.ConfigMapKeyRef != nil {
					queries = append(queries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Scope:  scope,
							Method: sdp.QueryMethod_GET,
							Query:  env.ValueFrom.ConfigMapKeyRef.Name,
							Type:   "ConfigMap",
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the config map could easily break the pod
							In: true,
							// The pod however isn't going to affect a config map
							Out: false,
						},
					})
				}
			}
		}

		for _, envFrom := range container.EnvFrom {
			if envFrom.SecretRef != nil {
				// Add linked item from spec.containers[].EnvFrom[].secretKeyRef
				queries = append(queries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Scope:  scope,
						Method: sdp.QueryMethod_GET,
						Query:  envFrom.SecretRef.Name,
						Type:   "Secret",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the secret could easily break the pod
						In: true,
						// The pod however isn't going to affect a secret
						Out: false,
					},
				})
			}
		}
	}

	if resource.Spec.PriorityClassName != "" {
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Scope:  sd.ClusterName,
				Method: sdp.QueryMethod_GET,
				Query:  resource.Spec.PriorityClassName,
				Type:   "PriorityClass",
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changing the priority class could break a pod by meaning that
				// it would now be scheduled with a lower priority and could
				// therefore end up pending for ages
				In: true,
				// The pod however isn't going to affect a priority class
				Out: false,
			},
		})
	}

	if len(resource.Status.PodIPs) > 0 {
		for _, ip := range resource.Status.PodIPs {
			queries = append(queries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Scope:  "global",
					Method: sdp.QueryMethod_GET,
					Query:  ip.IP,
					Type:   "ip",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// IPs go in both directions
					In:  true,
					Out: true,
				},
			})
		}
	} else if resource.Status.PodIP != "" {
		queries = append(queries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "ip",
				Method: sdp.QueryMethod_GET,
				Query:  resource.Status.PodIP,
				Scope:  "global",
			},
			BlastPropagation: &sdp.BlastPropagation{
				// IPs go in both directions
				In:  true,
				Out: true,
			},
		})
	}

	return queries, nil
}

func newPodSource(cs *kubernetes.Clientset, cluster string, namespaces []string) discovery.Source {
	return &KubeTypeSource[*v1.Pod, *v1.PodList]{
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

func init() {
	registerSourceLoader(newPodSource)
}
