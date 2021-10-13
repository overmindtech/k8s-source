package sources

import (
	"testing"
)

var secretYAML = `
apiVersion: v1
kind: Secret
metadata:
  name: secret-basic-auth
  namespace: k8s-source-testing
type: kubernetes.io/basic-auth
stringData:
  username: admin
  password: t0p-Secret
`

func TestSecretSource(t *testing.T) {
	var err error
	var source ResourceSource

	// Create the required pod
	err = CurrentCluster.Apply(secretYAML)

	t.Cleanup(func() {
		CurrentCluster.Delete(secretYAML)
	})

	if err != nil {
		t.Error(err)
	}

	source, err = SecretSource(CurrentCluster.ClientSet)

	if err != nil {
		t.Error(err)
	}

	BasicGetFindSearchTests(t, `{"fieldSelector": "metadata.name=secret-basic-auth"}`, source)
}
