package sources

import (
	"regexp"
	"testing"

	"github.com/overmindtech/sdp-go"
)

var serviceYAML = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: service-test-deployment
spec:
  selector:
    matchLabels:
      app: service-test
  replicas: 1
  template:
    metadata:
      labels:
        app: service-test
    spec:
      containers:
      - name: my-container
        image: nginx
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: service-test-service
spec:
  selector:
    app: service-test
  ports:
  - name: http
    protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
  externalName: service-test-external
`

func TestServiceSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := newServiceSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:    source,
		GetQuery:  "service-test-service",
		GetScope:  sd.String(),
		SetupYAML: serviceYAML,
		GetQueryTests: QueryTests{
			{
				ExpectedType:         "Pod",
				ExpectedMethod:       sdp.QueryMethod_SEARCH,
				ExpectedScope:        sd.String(),
				ExpectedQueryMatches: regexp.MustCompile(`app=service-test`),
			},
			{
				ExpectedType:   "Endpoint",
				ExpectedMethod: sdp.QueryMethod_GET,
				ExpectedQuery:  "service-test-service",
				ExpectedScope:  sd.String(),
			},
			{
				ExpectedType:   "dns",
				ExpectedMethod: sdp.QueryMethod_GET,
				ExpectedQuery:  "service-test-external",
				ExpectedScope:  "global",
			},
		},
	}

	st.Execute(t)
}
