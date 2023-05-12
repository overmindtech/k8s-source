package sources

import (
	"testing"
)

func TestNodeSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := NewNodeSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:        &source,
		GetQuery:      "k8s-source-tests-control-plane",
		GetScope:      sd.String(),
		GetQueryTests: QueryTests{},
	}

	st.Execute(t)
}
