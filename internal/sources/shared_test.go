package sources

import (
	"log"
	"os"
	"testing"

	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
)

type TestCluster struct {
	Name       string
	Kubeconfig string
	provider   *cluster.Provider
	T          *testing.T
}

func (t *TestCluster) Start() error {
	kubeconfigFile, err := os.CreateTemp("", "*-kubeconfig")

	if err != nil {
		return err
	}

	t.Name = "k8s-source-tests"
	t.Kubeconfig = kubeconfigFile.Name()

	t.provider = cluster.NewProvider()
	err = t.provider.Create(t.Name, cluster.CreateWithV1Alpha4Config(&v1alpha4.Cluster{}))

	if err != nil {
		return err
	}

	err = t.provider.ExportKubeConfig(t.Name, t.Kubeconfig)

	if err != nil {
		return err
	}

	return nil
}

func (t *TestCluster) Stop() error {
	err := t.provider.Delete(t.Name, t.Kubeconfig)

	os.Remove(t.Kubeconfig)

	return err
}

var CurrentCluster TestCluster

func TestMain(m *testing.M) {
	CurrentCluster = TestCluster{}

	err := CurrentCluster.Start()

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	code := m.Run()

	err = CurrentCluster.Stop()

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	os.Exit(code)
}
