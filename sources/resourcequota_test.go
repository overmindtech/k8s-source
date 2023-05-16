package sources

import (
	"testing"
)

var resourceQuotaYAML = `
apiVersion: v1
kind: ResourceQuota
metadata:
  name: quota-example
spec:
  hard:
    pods: "10"
    requests.cpu: "2"
    requests.memory: 2Gi
    limits.cpu: "4"
    limits.memory: 4Gi
`

func TestResourceQuotaSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := newResourceQuotaSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:    source,
		GetQuery:  "quota-example",
		GetScope:  sd.String(),
		SetupYAML: resourceQuotaYAML,
	}

	st.Execute(t)
}
