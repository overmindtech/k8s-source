package sources

import (
	"testing"
)

var RoleYAML = `
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: role-test-role
rules:
  - apiGroups:
      - ""
      - "apps"
      - "batch"
      - "extensions"
    resources:
      - pods
      - deployments
      - jobs
      - cronjobs
      - configmaps
      - secrets
    verbs:
      - get
      - list
      - watch
      - create
      - update
      - delete
`

func TestRoleSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := newRoleSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:    source,
		GetQuery:  "role-test-role",
		GetScope:  sd.String(),
		SetupYAML: RoleYAML,
	}

	st.Execute(t)
}
