package sources

import (
	"regexp"
	"testing"

	"github.com/overmindtech/sdp-go"
)

var NetworkPolicyYAML = `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-nginx
spec:
  podSelector:
    matchLabels:
      app: nginx
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: frontend
    ports:
    - protocol: TCP
      port: 80
`

func TestNetworkPolicySource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := newNetworkPolicySource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:    source,
		GetQuery:  "allow-nginx",
		GetScope:  sd.String(),
		SetupYAML: NetworkPolicyYAML,
		GetQueryTests: QueryTests{
			{
				ExpectedQueryMatches: regexp.MustCompile("nginx"),
				ExpectedType:         "Pod",
				ExpectedMethod:       sdp.QueryMethod_SEARCH,
				ExpectedScope:        sd.String(),
			},
		},
	}

	st.Execute(t)
}
