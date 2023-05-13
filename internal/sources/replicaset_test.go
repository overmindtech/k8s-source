package sources

import (
	"regexp"
	"testing"

	"github.com/overmindtech/sdp-go"
)

var replicaSetYAML = `
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: replica-set-test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: replica-set-test
  template:
    metadata:
      labels:
        app: replica-set-test
    spec:
      containers:
        - name: replica-set-test
          image: nginx:latest
          ports:
            - containerPort: 80

`

func TestReplicaSetSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := NewReplicaSetSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:    &source,
		GetQuery:  "replica-set-test",
		GetScope:  sd.String(),
		SetupYAML: replicaSetYAML,
		GetQueryTests: QueryTests{
			{
				ExpectedQueryMatches: regexp.MustCompile("app=replica-set-test"),
				ExpectedType:         "Pod",
				ExpectedMethod:       sdp.QueryMethod_SEARCH,
				ExpectedScope:        sd.String(),
			},
		},
	}

	st.Execute(t)
}
