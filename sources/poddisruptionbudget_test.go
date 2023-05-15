package sources

import (
	"regexp"
	"testing"

	"github.com/overmindtech/sdp-go"
)

var PodDisruptionBudgetYAML = `
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: example-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: example-app
`

func TestPodDisruptionBudgetSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := newPodDisruptionBudgetSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:    source,
		GetQuery:  "example-pdb",
		GetScope:  sd.String(),
		SetupYAML: PodDisruptionBudgetYAML,
		GetQueryTests: QueryTests{
			{
				ExpectedQueryMatches: regexp.MustCompile("app=example-app"),
				ExpectedType:         "Pod",
				ExpectedMethod:       sdp.QueryMethod_SEARCH,
				ExpectedScope:        sd.String(),
			},
		},
	}

	st.Execute(t)
}
