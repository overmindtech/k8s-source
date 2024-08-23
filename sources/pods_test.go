package sources

import (
	"regexp"
	"testing"

	"github.com/overmindtech/sdp-go"
)

var PodYAML = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pod-test-serviceaccount
---
apiVersion: v1
kind: Secret
metadata:
  name: pod-test-secret
type: Opaque
data:
  username: dXNlcm5hbWU=
  password: cGFzc3dvcmQ=
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: pod-test-configmap
data:
  config.ini: |
    [database]
    host=example.com
    port=5432
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: pod-test-configmap-cert
data:
  ca.pem: |
    wow such cert
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pod-test-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: Pod
metadata:
  name: pod-test-pod
spec:
  serviceAccountName: pod-test-serviceaccount
  volumes:
  - name: pod-test-pvc-volume
    persistentVolumeClaim:
      claimName: pod-test-pvc
  - name: database-config
    configMap:
      name: pod-test-configmap
  - name: projected-config
    projected:
      sources:
        - configMap:
            name: pod-test-configmap-cert
            items:
              - key: ca.pem
                path: ca.pem
  containers:
  - name: pod-test-container
    image: nginx
    volumeMounts:
    - name: pod-test-pvc-volume
      mountPath: /mnt/data
    - name: database-config
      mountPath: /etc/database
    - name: projected-config
      mountPath: /etc/projected
    envFrom:
    - secretRef:
        name: pod-test-secret
`

func TestPodSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := newPodSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:    source,
		GetQuery:  "pod-test-pod",
		GetScope:  sd.String(),
		SetupYAML: PodYAML,
		GetQueryTests: QueryTests{
			{
				ExpectedQueryMatches: regexp.MustCompile(`10\.`),
				ExpectedType:         "ip",
				ExpectedMethod:       sdp.QueryMethod_GET,
				ExpectedScope:        "global",
			},
			{
				ExpectedType:   "ServiceAccount",
				ExpectedMethod: sdp.QueryMethod_GET,
				ExpectedQuery:  "pod-test-serviceaccount",
				ExpectedScope:  sd.String(),
			},
			{
				ExpectedType:   "Secret",
				ExpectedMethod: sdp.QueryMethod_GET,
				ExpectedQuery:  "pod-test-secret",
				ExpectedScope:  sd.String(),
			},
			{
				ExpectedType:   "ConfigMap",
				ExpectedMethod: sdp.QueryMethod_GET,
				ExpectedQuery:  "pod-test-configmap",
				ExpectedScope:  sd.String(),
			},
			{
				ExpectedType:   "PersistentVolumeClaim",
				ExpectedMethod: sdp.QueryMethod_GET,
				ExpectedQuery:  "pod-test-pvc",
				ExpectedScope:  sd.String(),
			},
			{
				ExpectedType:   "ConfigMap",
				ExpectedMethod: sdp.QueryMethod_GET,
				ExpectedQuery:  "pod-test-configmap-cert",
				ExpectedScope:  sd.String(),
			},
		},
		Wait: func(item *sdp.Item) bool {
			return len(item.GetLinkedItemQueries()) >= 9
		},
	}

	st.Execute(t)
}
