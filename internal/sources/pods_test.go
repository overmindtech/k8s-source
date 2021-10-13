package sources

import (
	"fmt"
	"testing"
)

func TestPodSource(t *testing.T) {
	source, err := PodSource(CurrentCluster.ClientSet)

	if err != nil {
		t.Error(err)
	}

	items, err := source.Find("test.default")

	if err != nil {
		t.Error(err)
	}

	fmt.Println(items)
}
