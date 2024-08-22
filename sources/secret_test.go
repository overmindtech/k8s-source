package sources

import (
	"testing"
)

//nolint:gosec // this is just a test
var secretYAML = `
apiVersion: v1
kind: Secret
metadata:
  name: secret-test-secret
type: Opaque
data:
  username: dXNlcm5hbWUx   # base64-encoded "username1"
  password: cGFzc3dvcmQx   # base64-encoded "password1"

`

func TestSecretSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := newSecretSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:    source,
		GetQuery:  "secret-test-secret",
		GetScope:  sd.String(),
		SetupYAML: secretYAML,
	}

	st.Execute(t)
}
