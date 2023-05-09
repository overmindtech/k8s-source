// Reusable testing libraries for testing backends
package sources

import (
	"context"
	"testing"
	"time"

	"github.com/overmindtech/sdp-go"
)

// BasicGetListSearchTests Executes a series of basic tests against a kubernetes
// source. These tests include:
//
//   - Searching with a given query for items
//   - Grabbing the name of one of the found items and making sure that we can
//     Get() it
//   - Running a List() and ensuring that there is > 0 results
func BasicGetListSearchTests(t *testing.T, query string, source ResourceSource) {
	itemScope := "testcluster:443.k8s-source-testing"

	var getName string

	t.Run("Testing basic Search()", func(t *testing.T) {
		var items []*sdp.Item
		var err error

		// Give it some time for the pod to come up
		for i := 0; i < 30; i++ {
			items, err = source.Search(context.Background(), itemScope, query)

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

		item, err := source.Get(context.Background(), itemScope, getName)

		if err != nil {
			t.Error(err)
		}

		if x := item.UniqueAttributeValue(); x != getName {
			t.Errorf("expected pod name hello, got %v", x)
		}

		TestValidateItem(t, item)
	})

	t.Run("Testing basic List()", func(t *testing.T) {
		if getName == "" {
			t.Skip("Nothing found from Search(), skipping")
		}

		items, err := source.List(context.Background(), itemScope)

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
			t.Fatalf("List() did not return pod %v", getName)
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

	if i.GetScope() == "" {
		t.Errorf("Item %v has an empty Scope", i.GetScope())
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

		// We don't need to check for an empty scope here since if it's empty
		// it will just inherit the scope of the parent
	}

	for index, linkedQuery := range i.GetLinkedItemQueries() {
		if linkedQuery.GetType() == "" {
			t.Errorf("LinkedQuery %v of item %v has empty type", index, i.GloballyUniqueName())

		}

		if linkedQuery.GetMethod() != sdp.QueryMethod_LIST {
			if linkedQuery.GetQuery() == "" {
				t.Errorf("LinkedQuery %v of item %v has empty query. This is not allowed unless the method is LIST", index, i.GloballyUniqueName())
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
