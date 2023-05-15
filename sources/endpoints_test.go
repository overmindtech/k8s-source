package sources

import (
	"regexp"
	"testing"

	"github.com/overmindtech/sdp-go"
)

var endpointsYAML = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: endpoint-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: endpoint-test
  template:
    metadata:
      labels:
        app: endpoint-test
    spec:
      containers:
        - name: endpoint-test
          image: nginx:latest
          ports:
            - containerPort: 80

---
apiVersion: v1
kind: Service
metadata:
  name: endpoint-service
spec:
  selector:
    app: endpoint-test
  ports:
    - name: http
      port: 80
      targetPort: 80
  type: ClusterIP

`

func TestEndpointsSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := newEndpointsSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:    source,
		GetQuery:  "endpoint-service",
		GetScope:  sd.String(),
		SetupYAML: endpointsYAML,
		GetQueryTests: QueryTests{
			{
				ExpectedQueryMatches: regexp.MustCompile(`^10\.`),
				ExpectedType:         "ip",
				ExpectedMethod:       sdp.QueryMethod_GET,
				ExpectedScope:        "global",
			},
			{
				ExpectedType:   "Node",
				ExpectedMethod: sdp.QueryMethod_GET,
				ExpectedQuery:  "k8s-source-tests-control-plane",
				ExpectedScope:  CurrentCluster.Name,
			},
			{
				ExpectedType:         "Pod",
				ExpectedMethod:       sdp.QueryMethod_GET,
				ExpectedQueryMatches: regexp.MustCompile("endpoint-deployment"),
				ExpectedScope:        sd.String(),
			},
		},
		Wait: func(item *sdp.Item) bool {
			return len(item.LinkedItemQueries) > 0
		},
	}

	st.Execute(t)
}
