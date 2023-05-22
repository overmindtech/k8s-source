package sources

import (
	"testing"
)

var clusterRoleYAML = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: read-only
rules:
- apiGroups: [""]
  resources: ["*"]
  verbs: ["get", "list", "watch"]

`

func TestClusterRoleSource(t *testing.T) {
	source := newClusterRoleSource(CurrentCluster.ClientSet, CurrentCluster.Name, []string{})

	st := SourceTests{
		Source:        source,
		GetQuery:      "read-only",
		GetScope:      CurrentCluster.Name,
		SetupYAML:     clusterRoleYAML,
		GetQueryTests: QueryTests{},
	}

	st.Execute(t)
}
