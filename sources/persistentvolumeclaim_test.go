package sources

import (
	"testing"
)

var persistentVolumeClaimYAML = `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc-test-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pvc-test-pv
spec:
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: /tmp/pvc-test-pv
---
apiVersion: v1
kind: Pod
metadata:
  name: pvc-test-pod
spec:
  containers:
  - name: pvc-test-container
    image: nginx
    volumeMounts:
    - name: pvc-test-volume
      mountPath: /data
  volumes:
  - name: pvc-test-volume
    persistentVolumeClaim:
      claimName: pvc-test-pvc
`

func TestPersistentVolumeClaimSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := newPersistentVolumeClaimSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:        source,
		GetQuery:      "pvc-test-pvc",
		GetScope:      sd.String(),
		SetupYAML:     persistentVolumeClaimYAML,
		GetQueryTests: QueryTests{},
	}

	st.Execute(t)
}
