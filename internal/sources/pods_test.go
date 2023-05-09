package sources

import (
	"testing"
)

var podYAML = `
apiVersion: batch/v1
kind: Job
metadata:
  name: hello
  namespace: k8s-source-testing
spec:
  template:
    # This is the pod template
    spec:
      containers:
      - name: hello
        image: busybox
        command: ['sh', '-c', 'echo "Hello, Kubernetes!" && sleep 3600']
      restartPolicy: OnFailure
    # The pod template ends here
`

func TestPodSource(t *testing.T) {
	var err error
	var source ResourceSource

	// Create the required pod
	err = CurrentCluster.Apply(podYAML)

	t.Cleanup(func() {
		CurrentCluster.Delete(podYAML)
	})

	if err != nil {
		t.Error(err)
	}

	source, err = PodSource(CurrentCluster.ClientSet)

	if err != nil {
		t.Error(err)
	}

	BasicGetListSearchTests(t, `{"labelSelector": "job-name=hello"}`, source)
}
