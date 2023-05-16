package sources

import (
	"testing"
)

var priorityClassYAML = `
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: ultra-mega-priority
value: 1000000
globalDefault: false
description: "This priority class should be used for ultra-mega-priority workloads"
`

func TestPriorityClassSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := newPriorityClassSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:        source,
		GetQuery:      "ultra-mega-priority",
		GetScope:      sd.String(),
		SetupYAML:     priorityClassYAML,
		GetQueryTests: QueryTests{},
	}

	st.Execute(t)
}
