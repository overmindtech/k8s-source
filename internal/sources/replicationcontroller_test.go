package sources

import (
	"regexp"
	"testing"

	"github.com/overmindtech/sdp-go"
)

var replicationControllerYAML = `
apiVersion: v1
kind: ReplicationController
metadata:
  name: replication-controller-test
spec:
  replicas: 1
  selector:
    app: replication-controller-test
  template:
    metadata:
      labels:
        app: replication-controller-test
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80

`

func TestReplicationControllerSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := NewReplicationControllerSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:    &source,
		GetQuery:  "replication-controller-test",
		GetScope:  sd.String(),
		SetupYAML: replicationControllerYAML,
		GetQueryTests: QueryTests{
			{
				ExpectedQueryMatches: regexp.MustCompile("app=replication-controller-test"),
				ExpectedType:         "Pod",
				ExpectedMethod:       sdp.QueryMethod_SEARCH,
				ExpectedScope:        sd.String(),
			},
		},
	}

	st.Execute(t)
}
