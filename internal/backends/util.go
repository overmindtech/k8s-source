package backends

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/dylanratcliffe/sdp-go"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Selector represents a set of key value pairs that we are going to use as a
// selector
type Selector map[string]string

// String converts a set of key value pairs to the string format that a selector
// is expecting
func (l Selector) String() string {
	var conditions []string

	conditions = make([]string, 0)

	for k, v := range l {
		conditions = append(conditions, fmt.Sprintf("%v=%v", k, v))
	}

	return strings.Join(conditions, ",")
}

// ThroughJSON Converts the object though JSON and back returning an interface
// with all other type data stripped
func ThroughJSON(i interface{}) (interface{}, error) {
	var ri interface{}
	var jsonData []byte
	var err error

	// Marshall to JSON
	if jsonData, err = json.Marshal(i); err != nil {
		return ri, err
	}

	// Convert back
	err = json.Unmarshal(jsonData, &ri)

	return ri, err
}

// GetK8sMeta will assign attributes to an existing attributes hash that are
// taken from Kubernetes ObjectMetadata
func GetK8sMeta(s metaV1.Object) map[string]interface{} {
	a := make(map[string]interface{})

	// TODO: I could do this with even more dynamic reflection...

	if v := reflect.ValueOf(s.GetNamespace()); !v.IsZero() {
		a["namespace"] = s.GetNamespace()
	}

	if v := reflect.ValueOf(s.GetName()); !v.IsZero() {
		a["name"] = s.GetName()
	}

	if v := reflect.ValueOf(s.GetGenerateName()); !v.IsZero() {
		a["generateName"] = s.GetGenerateName()
	}

	if v := reflect.ValueOf(s.GetUID()); !v.IsZero() {
		a["uID"] = s.GetUID()
	}

	if v := reflect.ValueOf(s.GetResourceVersion()); !v.IsZero() {
		a["resourceVersion"] = s.GetResourceVersion()
	}

	if v := reflect.ValueOf(s.GetGeneration()); !v.IsZero() {
		a["generation"] = s.GetGeneration()
	}

	if v := reflect.ValueOf(s.GetSelfLink()); !v.IsZero() {
		a["selfLink"] = s.GetSelfLink()
	}

	if v := reflect.ValueOf(s.GetCreationTimestamp()); !v.IsZero() {
		a["creationTimestamp"] = s.GetCreationTimestamp()
	}

	if v := reflect.ValueOf(s.GetDeletionTimestamp()); !v.IsZero() {
		a["deletionTimestamp"] = s.GetDeletionTimestamp()
	}

	if v := reflect.ValueOf(s.GetDeletionGracePeriodSeconds()); !v.IsZero() {
		a["deletionGracePeriodSeconds"] = s.GetDeletionGracePeriodSeconds()
	}

	if v := reflect.ValueOf(s.GetLabels()); !v.IsZero() {
		a["labels"] = s.GetLabels()
	}

	if v := reflect.ValueOf(s.GetAnnotations()); !v.IsZero() {
		a["annotations"] = s.GetAnnotations()
	}

	if v := reflect.ValueOf(s.GetFinalizers()); !v.IsZero() {
		a["finalizers"] = s.GetFinalizers()
	}

	if v := reflect.ValueOf(s.GetOwnerReferences()); !v.IsZero() {
		a["ownerReferences"] = s.GetOwnerReferences()
	}

	if v := reflect.ValueOf(s.GetClusterName()); !v.IsZero() {
		a["clusterName"] = s.GetClusterName()
	}

	// Note that we are deliberately ignoring ManagedFields here since it's a
	// lot of data and I'm not sure if its value

	return a
}

func ListOptionsToQuery(lo *metaV1.ListOptions) string {
	jsonData, err := json.Marshal(lo)

	if err == nil {
		return string(jsonData)
	}

	return ""
}

// SelectorToQuery converts a LabelSelector to JSON so that it can be passed to
// a SEARCH query

// TODO: Rename to LabelSelectorToQuery
func LabelSelectorToQuery(labelSelector *metaV1.LabelSelector) string {
	return ListOptionsToQuery(&metaV1.ListOptions{
		LabelSelector: Selector(labelSelector.MatchLabels).String(),
	})
}

// QueryToListOptions converts a Search() query string to a ListOptions object that can
// be used to query the API
func QueryToListOptions(query string) (metaV1.ListOptions, error) {
	var queryBytes []byte
	var err error
	var listOptions metaV1.ListOptions

	queryBytes = []byte(query)

	// Convert from JSON
	if err = json.Unmarshal(queryBytes, &listOptions); err != nil {
		return listOptions, err
	}

	// Override some of the things we don't want people to set
	listOptions.Watch = false

	return listOptions, nil
}

// ObjectReferenceToLIR Converts a K8s ObjectReference to a linked item request.
// Note that you must provide the parent context (the name of the cluster) since
// the reference could be an object in a different namespace and therefore
// context. If the parent context is empty, the context will be assumed to be
// the same as the current object
func ObjectReferenceToLIR(ref *coreV1.ObjectReference, parentContext string) *sdp.ItemRequest {
	if ref == nil {
		return nil
	}

	var context string

	// If we have a namespace then calculate the full context name
	if ref.Namespace != "" && parentContext != "" {
		context = fmt.Sprintf("%v.%v", parentContext, ref.Namespace)
	}

	return &sdp.ItemRequest{
		Type:    strings.ToLower(ref.Kind), // Lowercase as per convention
		Method:  sdp.RequestMethod_GET,     // Object references are to a specific object
		Query:   ref.Name,
		Context: context,
	}
}

// mapK8sObject Converts a kubernetes object in to an item that we understand.
// This includes extracting certain useful metadata using GetK8sMeta() and then
// extracting all other useful information, with some exceptions
func mapK8sObject(typ string, object metaV1.Object) (*sdp.Item, error) {
	var attributes map[string]interface{}
	var item *sdp.Item
	var v reflect.Value
	var t reflect.Type
	var err error

	item = &sdp.Item{
		Type:            typ,
		UniqueAttribute: "name",
	}
	attributes = GetK8sMeta(object)

	// Get the reflected details of the object
	v = reflect.Indirect(reflect.ValueOf(object))
	t = v.Type()

	// Ignore the following fields
	ignore := []string{
		"TypeMeta",
		"ObjectMeta",
		"BinaryData",
		"StringData",
		"Immutable",
	}

	if v.Kind() == reflect.Struct {
		// Range over fields
		n := t.NumField()
		for i := 0; i < n; i++ {
			field := t.Field(i)

			// Check if the field is meant to be ignored
			var shouldIgnore bool
			for _, f := range ignore {
				if f == field.Name {
					shouldIgnore = true
				}
			}
			if shouldIgnore {
				continue
			}

			// Get the zero value for this field
			zero := reflect.Zero(reflect.TypeOf(field)).Interface()

			// Check if the field is it's nil value
			// Check if there actually was a field with that name
			if !reflect.DeepEqual(field, zero) {
				attributes[strings.ToLower(field.Name)] = v.Field(i).Interface()
			}
		}
	}

	item.Attributes, err = sdp.ToAttributes(attributes)
	item.LinkedItemRequests = []*sdp.ItemRequest{}

	return item, err
}
