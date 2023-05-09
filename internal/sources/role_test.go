package sources

import (
	"testing"
)

var roleYAML = `
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: k8s-source-testing
  name: pod-reader
rules:
- apiGroups: [""] # "" indicates the core API group
  resources: ["pods"]
  verbs: ["get", "watch", "list"]
`

func TestRoleSource(t *testing.T) {
	var err error
	var source ResourceSource

	// Create the required pod
	err = CurrentCluster.Apply(roleYAML)

	t.Cleanup(func() {
		CurrentCluster.Delete(roleYAML)
	})

	if err != nil {
		t.Error(err)
	}

	source, err = RoleSource(CurrentCluster.ClientSet)

	if err != nil {
		t.Error(err)
	}

	BasicGetListSearchTests(t, `{"fieldSelector": "metadata.name=pod-reader"}`, source)
}
