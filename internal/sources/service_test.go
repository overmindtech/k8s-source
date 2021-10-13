package sources

import (
	"testing"
)

var serviceYAML = `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: k8s-source-testing
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
  namespace: k8s-source-testing
spec:
  selector:
    app: nginx
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
`

func TestServiceSource(t *testing.T) {
	var err error
	var source ResourceSource

	// Create the required pod
	err = CurrentCluster.Apply(serviceYAML)

	t.Cleanup(func() {
		CurrentCluster.Delete(serviceYAML)
	})

	if err != nil {
		t.Error(err)
	}

	source, err = ServiceSource(CurrentCluster.ClientSet)

	if err != nil {
		t.Error(err)
	}

	BasicGetFindSearchTests(t, `{"fieldSelector": "metadata.name=nginx-service"}`, source)
}
