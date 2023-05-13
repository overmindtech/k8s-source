package sources

import (
	"context"
	"errors"
	"fmt"

	"github.com/overmindtech/sdp-go"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespacedInterfaceBuilder The function that create a client to query a
// namespaced resource. e.g. `CoreV1().Pods`
type NamespacedInterfaceBuilder[Resource metav1.Object, ResourceList any] func(namespace string) ItemInterface[Resource, ResourceList]

// ClusterInterfaceBuilder The function that create a client to query a
// cluster-wide resource. e.g. `CoreV1().Nodes`
type ClusterInterfaceBuilder[Resource metav1.Object, ResourceList any] func() ItemInterface[Resource, ResourceList]

// ItemInterface An interface that matches the `Get` and `List` methods for K8s
// resources since these are the ones that we use for getting Overmind data.
// Kube's clients are usually namespaced when they are created, so this
// interface is expected to only returns items from a single namespace
type ItemInterface[Resource metav1.Object, ResourceList any] interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (Resource, error)
	List(ctx context.Context, opts metav1.ListOptions) (ResourceList, error)
}

type KubeTypeSource[Resource metav1.Object, ResourceList any] struct {
	// The function that creates a client to query a namespaced resource. e.g.
	// `CoreV1().Pods`. Either this or `NamespacedInterfaceBuilder` must be
	// specified
	ClusterInterfaceBuilder ClusterInterfaceBuilder[Resource, ResourceList]

	// The function that creates a client to query a cluster-wide resource. e.g.
	// `CoreV1().Nodes`. Either this or `ClusterInterfaceBuilder` must be
	// specified
	NamespacedInterfaceBuilder NamespacedInterfaceBuilder[Resource, ResourceList]

	// A function that extracts a slice of Resources from a ResourceList
	ListExtractor func(ResourceList) ([]Resource, error)

	// A function that returns a list of linked item queries for a given
	// resource and scope
	LinkedItemQueryExtractor func(resource Resource, scope string) ([]*sdp.Query, error)

	// A function that extracts health from the resource, this is optional
	HealthExtractor func(resource Resource) *sdp.Health

	// A function that redacts sensitive data from the resource, this is
	// optional
	Redact func(resource Resource) Resource

	// The type of items that this source should return. This should be the
	// "Kind" of the kubernetes resources, e.g. "Pod", "Node", "ServiceAccount"
	TypeName string
	// List of namespaces that this source should query
	Namespaces []string
	// The name of the cluster that this source is for. This is used to generate
	// scopes
	ClusterName string
}

// validate Validates that the source is correctly set up
func (k *KubeTypeSource[Resource, ResourceList]) Validate() error {
	if k.NamespacedInterfaceBuilder == nil && k.ClusterInterfaceBuilder == nil {
		return fmt.Errorf("either NamespacedInterfaceBuilder or ClusterInterfaceBuilder must be specified")
	}

	if k.ListExtractor == nil {
		return fmt.Errorf("ListExtractor must be specified")
	}

	if k.TypeName == "" {
		return fmt.Errorf("TypeName must be specified")
	}

	if k.namespaced() && len(k.Namespaces) == 0 {
		return fmt.Errorf("Namespaces must be specified when NamespacedInterfaceBuilder is specified")
	}

	if k.ClusterName == "" {
		return fmt.Errorf("ClusterName must be specified")
	}

	return nil
}

// namespaced Returns whether the source is namespaced or not
func (k *KubeTypeSource[Resource, ResourceList]) namespaced() bool {
	return k.NamespacedInterfaceBuilder != nil
}

func (k *KubeTypeSource[Resource, ResourceList]) Type() string {
	return k.TypeName
}

func (k *KubeTypeSource[Resource, ResourceList]) Name() string {
	return fmt.Sprintf("k8s-%v", k.TypeName)
}

func (k *KubeTypeSource[Resource, ResourceList]) Weight() int {
	return 10
}

func (k *KubeTypeSource[Resource, ResourceList]) Scopes() []string {
	namespaces := make([]string, 0)

	if k.namespaced() {
		for _, ns := range k.Namespaces {
			sd := ScopeDetails{
				ClusterName: k.ClusterName,
				Namespace:   ns,
			}

			namespaces = append(namespaces, sd.String())
		}
	} else {
		sd := ScopeDetails{
			ClusterName: k.ClusterName,
		}

		namespaces = append(namespaces, sd.String())
	}

	return namespaces
}

func (k *KubeTypeSource[Resource, ResourceList]) Get(ctx context.Context, scope string, query string) (*sdp.Item, error) {
	i, err := k.itemInterface(scope)

	if err != nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: err.Error(),
		}
	}

	resource, err := i.Get(ctx, query, metav1.GetOptions{})

	if err != nil {
		statusErr := new(k8serr.StatusError)

		if errors.As(err, &statusErr) && statusErr.ErrStatus.Code == 404 {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: statusErr.ErrStatus.Message,
			}
		}

		return nil, err
	}

	item, err := k.resourceToItem(resource)

	if err != nil {
		return nil, err
	}

	return item, nil
}

func (k *KubeTypeSource[Resource, ResourceList]) List(ctx context.Context, scope string) ([]*sdp.Item, error) {
	return k.listWithOptions(ctx, scope, metav1.ListOptions{})
}

// listWithOptions Runs the inbuilt list method with the given options
func (k *KubeTypeSource[Resource, ResourceList]) listWithOptions(ctx context.Context, scope string, opts metav1.ListOptions) ([]*sdp.Item, error) {
	i, err := k.itemInterface(scope)

	if err != nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOSCOPE,
			ErrorString: err.Error(),
		}
	}

	list, err := i.List(ctx, opts)

	if err != nil {
		return nil, err
	}

	resourceList, err := k.ListExtractor(list)

	if err != nil {
		return nil, err
	}

	items, err := k.resourcesToItems(resourceList, scope)

	if err != nil {
		return nil, err
	}

	return items, nil
}

func (k *KubeTypeSource[Resource, ResourceList]) Search(ctx context.Context, scope string, query string) ([]*sdp.Item, error) {
	opts, err := QueryToListOptions(query)

	if err != nil {
		return nil, err
	}

	return k.listWithOptions(ctx, scope, opts)
}

// itemInterface Returns the correct interface depending on whether the source
// is namespaced or not
func (k *KubeTypeSource[Resource, ResourceList]) itemInterface(scope string) (ItemInterface[Resource, ResourceList], error) {
	// If this is a namespaced resource, then parse the scope to get the
	// namespace
	if k.namespaced() {
		details, err := ParseScope(scope, k.namespaced())

		if err != nil {
			return nil, err
		}

		return k.NamespacedInterfaceBuilder(details.Namespace), nil
	} else {
		return k.ClusterInterfaceBuilder(), nil
	}
}

var ignoredMetadataFields = []string{
	"managedFields",
	"binaryData",
	"immutable",
	"stringData",
}

func ignored(key string) bool {
	for _, ignoredKey := range ignoredMetadataFields {
		if key == ignoredKey {
			return true
		}
	}

	return false
}

// resourcesToItems Converts a slice of resources to a slice of items
func (k *KubeTypeSource[Resource, ResourceList]) resourcesToItems(resourceList []Resource, scope string) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, len(resourceList))

	var err error

	for i := range resourceList {
		items[i], err = k.resourceToItem(resourceList[i])

		if err != nil {
			return nil, err
		}

	}

	return items, nil
}

// resourceToItem Converts a resource to an item
func (k *KubeTypeSource[Resource, ResourceList]) resourceToItem(resource Resource) (*sdp.Item, error) {
	sd := ScopeDetails{
		ClusterName: k.ClusterName,
		Namespace:   resource.GetNamespace(),
	}

	// Redact sensitive data if required
	if k.Redact != nil {
		resource = k.Redact(resource)
	}

	attributes, err := sdp.ToAttributesViaJson(resource)

	if err != nil {
		return nil, err
	}

	// Promote the metadata to the top level
	if metadata, err := attributes.Get("metadata"); err == nil {
		// Cast to a type we can iterate over
		if metadataMap, ok := metadata.(map[string]interface{}); ok {
			for key, value := range metadataMap {
				// Check that the key isn't in the ignored list
				if !ignored(key) {
					attributes.Set(key, value)
				}
			}
		}

		// Remove the metadata attribute
		attributes.AttrStruct.Fields["metadata"] = nil
	}

	item := &sdp.Item{
		Type:            resource.GetName(),
		UniqueAttribute: "name",
		Scope:           sd.String(),
		Attributes:      attributes,
	}

	// Automatically create links to owner references
	for _, ref := range resource.GetOwnerReferences() {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
			Type:   ref.Kind,
			Method: sdp.QueryMethod_GET,
			Query:  ref.Name,
			Scope:  sd.String(),
		})
	}

	if k.LinkedItemQueryExtractor != nil {
		// Add linked items
		newQueries, err := k.LinkedItemQueryExtractor(resource, sd.String())

		if err != nil {
			return nil, err
		}

		item.LinkedItemQueries = append(item.LinkedItemQueries, newQueries...)
	}

	if k.HealthExtractor != nil {
		item.Health = k.HealthExtractor(resource)
	}

	return item, nil
}

// ObjectReferenceToQuery Converts a K8s ObjectReference to a linked item
// request. Note that you must provide the parent scope since the reference
// could be an object in a different namespace, if it is we need to re-use the
// cluster name from the parent scope
func ObjectReferenceToQuery(ref *corev1.ObjectReference, parentScope ScopeDetails) *sdp.Query {
	if ref == nil {
		return nil
	}

	// Update the namespace, but keep the cluster the same
	parentScope.Namespace = ref.Namespace

	return &sdp.Query{
		Type:   ref.Kind,
		Method: sdp.QueryMethod_GET, // Object references are to a specific object
		Query:  ref.Name,
		Scope:  parentScope.String(),
	}
}
