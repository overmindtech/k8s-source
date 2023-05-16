package sources

import (
	"testing"
)

var deploymentYAML = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-deployment
  template:
    metadata:
      labels:
        app: my-deployment
    spec:
      containers:
      - name: my-container
        image: nginx:latest
        ports:
        - containerPort: 80
`

func TestDeploymentSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := newDeploymentSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:    source,
		GetQuery:  "my-deployment",
		GetScope:  sd.String(),
		SetupYAML: deploymentYAML,
	}

	st.Execute(t)
}
