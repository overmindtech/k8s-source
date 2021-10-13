// Reusable testing libraries for testing backends
package sources

import (
	"testing"
	"time"

	"github.com/dylanratcliffe/sdp-go"
)

// BasicGetFindSearchTests Executes a series of basic tests against a kubernetes
// source. These tests include:
//
// * Searching with a given query for items
// * Grabbing the name of one of the found items and making sure that we can
//   Get() it
// * Running a Find() and ensuring that there is > 0 results
func BasicGetFindSearchTests(t *testing.T, query string, source ResourceSource) {
	itemContext := "testcluster:443.k8s-source-testing"

	var getName string

	t.Run("Testing basic Search()", func(t *testing.T) {
		var items []*sdp.Item
		var err error

		// Give it some time for the pod to come up
		for i := 0; i < 30; i++ {
			items, err = source.Search(itemContext, query)

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

	t.Run("Testing basic Get()", func(t *testing.T) {
		if getName == "" {
			t.Skip("Nothing found from Search(), skipping")
		}

		item, err := source.Get(itemContext, getName)

		if err != nil {
			t.Error(err)
		}

		if x := item.UniqueAttributeValue(); x != getName {
			t.Errorf("expected pod name hello, got %v", x)
		}

		TestValidateItem(t, item)
	})

	t.Run("Testing basic Find()", func(t *testing.T) {
		if getName == "" {
			t.Skip("Nothing found from Search(), skipping")
		}

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

// TestValidateItem Checks an item to ensure it is a valid SDP item. This includes
// checking that all required attributes are populated
func TestValidateItem(t *testing.T, i *sdp.Item) {
	// Ensure that the item has the required fields set i.e.
	//
	// * Type
	// * UniqueAttribute
	// * Attributes
	if i.GetType() == "" {
		t.Errorf("Item %v has an empty Type", i.GloballyUniqueName())
	}

	if i.GetUniqueAttribute() == "" {
		t.Errorf("Item %v has an empty UniqueAttribute", i.GloballyUniqueName())
	}

	if i.GetContext() == "" {
		t.Errorf("Item %v has an empty Context", i.GetContext())
	}

	attrMap := i.GetAttributes().AttrStruct.AsMap()

	if len(attrMap) == 0 {
		t.Errorf("Attributes for item %v are empty", i.GloballyUniqueName())
	}

	// Check the attributes themselves for validity
	for k := range attrMap {
		if k == "" {
			t.Errorf("Item %v has an attribute with an empty name", i.GloballyUniqueName())
		}
	}

	// Make sure that the UniqueAttributeValue is populated
	if i.UniqueAttributeValue() == "" {
		t.Errorf("UniqueAttribute %v for item %v is empty", i.GetUniqueAttribute(), i.GloballyUniqueName())
	}

	for index, linkedItem := range i.GetLinkedItems() {
		if linkedItem.GetType() == "" {
			t.Errorf("LinkedItem %v of item %v has empty type", index, i.GloballyUniqueName())
		}

		if linkedItem.GetUniqueAttributeValue() == "" {
			t.Errorf("LinkedItem %v of item %v has empty UniqueAttributeValue", index, i.GloballyUniqueName())
		}

		// We don't need to check for an empty context here since if it's empty
		// it will just inherit the context of the parent
	}

	for index, linkedItemRequest := range i.GetLinkedItemRequests() {
		if linkedItemRequest.GetType() == "" {
			t.Errorf("LinkedItemRequest %v of item %v has empty type", index, i.GloballyUniqueName())

		}

		if linkedItemRequest.GetMethod() != sdp.RequestMethod_FIND {
			if linkedItemRequest.GetQuery() == "" {
				t.Errorf("LinkedItemRequest %v of item %v has empty query. This is not allowed unless the method is FIND", index, i.GloballyUniqueName())
			}
		}
	}
}

// TestValidateItems Runs TestValidateItem on many items
func TestValidateItems(t *testing.T, is []*sdp.Item) {
	for _, i := range is {
		TestValidateItem(t, i)
	}
}
