package sources

import (
	"testing"
	"time"

	"github.com/dylanratcliffe/sdp-go"
)

var podYAML = `
apiVersion: batch/v1
kind: Job
metadata:
  name: hello
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

	itemContext := "testcluster:443.default"

	var getName string

	t.Run("Testing Search()", func(t *testing.T) {
		var items []*sdp.Item
		var err error

		// Give it some time for the pod to come up
		for i := 0; i < 5; i++ {
			items, err = source.Search(itemContext, `{"labelSelector": "job-name=hello"}`)

			if len(items) > 0 {
				break
			}

			time.Sleep(1 * time.Second)
		}

		if err != nil {
			t.Fatal(err)
		}

		if l := len(items); l != 1 {
			t.Fatalf("Expected 1 item, got %v", l)
		}

		item := items[0]

		TestValidateItem(t, item)

		// Populate this so that the get test can work
		getName = item.UniqueAttributeValue()
	})

	t.Run("Testing Get()", func(t *testing.T) {
		// TODO: find out what the name of the pod is first
		item, err := source.Get(itemContext, getName)

		if err != nil {
			t.Error(err)
		}

		if x := item.UniqueAttributeValue(); x != getName {
			t.Errorf("expected pod name hello, got %v", x)
		}

		TestValidateItem(t, item)
	})

	t.Run("Testing Find()", func(t *testing.T) {
		items, err := source.Find(itemContext)

		if err != nil {
			t.Fatal(err)
		}

		found := false

		// Make sure the item from the get was there
		for _, item := range items {
			TestValidateItem(t, item)

			if item.UniqueAttributeValue() == getName {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("Find() did not return pod %v", getName)
		}
	})

}
