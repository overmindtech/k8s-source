package sources

import (
	"regexp"
	"testing"

	"github.com/overmindtech/sdp-go"
)

func TestNodeSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := NewNodeSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:   &source,
		GetQuery: "k8s-source-tests-control-plane",
		GetScope: sd.String(),
		GetQueryTests: QueryTests{
			{
				ExpectedType:         "ip",
				ExpectedMethod:       sdp.QueryMethod_GET,
				ExpectedScope:        "global",
				ExpectedQueryMatches: regexp.MustCompile(`172\.`),
			},
		},
	}

	st.Execute(t)
}
