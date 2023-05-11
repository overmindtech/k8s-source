package sources

import "testing"

var configMapYAML = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-configmap
data:
  DATABASE_URL: "postgres://myuser:mypassword@mydbhost:5432/mydatabase"
  APP_SECRET_KEY: "mysecretkey123"
`

func TestConfigMapSource(t *testing.T) {
	sd := ScopeDetails{
		ClusterName: CurrentCluster.Name,
		Namespace:   "default",
	}

	source := NewConfigMapSource(CurrentCluster.ClientSet, sd.ClusterName, []string{sd.Namespace})

	st := SourceTests{
		Source:        &source,
		GetQuery:      "my-configmap",
		GetScope:      sd.String(),
		SetupYAML:     configMapYAML,
		GetQueryTests: QueryTests{},
	}

	st.Execute(t)
}
