package sources

import "testing"

func TestParseScope(t *testing.T) {
	type ParseTest struct {
		Input       string
		ClusterName string
		Namespace   string
	}

	tests := []ParseTest{
		{
			Input:       "127.0.0.1:61081.default",
			ClusterName: "127.0.0.1:61081",
			Namespace:   "default",
		},
		{
			Input:       "127.0.0.1:61081.kube-node-lease",
			ClusterName: "127.0.0.1:61081",
			Namespace:   "kube-node-lease",
		},
		{
			Input:       "127.0.0.1:61081.kube-public",
			ClusterName: "127.0.0.1:61081",
			Namespace:   "kube-public",
		},
		{
			Input:       "127.0.0.1:61081.kube-system",
			ClusterName: "127.0.0.1:61081",
			Namespace:   "kube-system",
		},
		{
			Input:       "127.0.0.1:61081",
			ClusterName: "127.0.0.1:61081",
			Namespace:   "",
		},
		{
			Input:       "cluster1.k8s.company.com:443",
			ClusterName: "cluster1.k8s.company.com:443",
			Namespace:   "",
		},
		{
			Input:       "cluster1.k8s.company.com",
			ClusterName: "cluster1.k8s.company.com",
			Namespace:   "",
		},
		{
			Input:       "test",
			ClusterName: "test",
			Namespace:   "",
		},
	}

	for _, test := range tests {
		result := ParseScope(test.Input)

		if test.ClusterName != result.ClusterName {
			t.Errorf("ClusterName did not match, expected %v, got %v", test.ClusterName, result.ClusterName)
		}

		if test.Namespace != result.Namespace {
			t.Errorf("Namespace did not match, expected %v, got %v", test.Namespace, result.Namespace)
		}
	}

}
