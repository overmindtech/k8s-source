package sources

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/overmindtech/sdp-go"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	log "github.com/sirupsen/logrus"
)

// defaultAPITimeout is the default amount of time to wait per for each API
// query to K8s. This is passed as a context to API request functions
var apiTimeoutDefault = (10 * time.Second)
var apiTimeoutSet = false
var apiTimeout time.Duration

// ClusterName stores the name of the cluster, this is also used as the scope
// for non-namespaced items. This designed to be user by namespaced items to
// create linked item requests on non-namespaced items
var ClusterName string

// NamespacedSourceFunction is a function that accepts a kubernetes client and
// namespace, and returns a ResourceSource for a given type. This also satisfies
// the Backend interface
type NamespacedSourceFunction func(cs *kubernetes.Clientset) ResourceSource

// NonNamespacedSourceFunction is a function that accepts a kubernetes client and
// returns a ResourceSource for a given type. This also satisfies the Backend
// interface
type NonNamespacedSourceFunction func(cs *kubernetes.Clientset) ResourceSource

// SourceFunction is a function that accepts a kubernetes client and returns a
// ResourceSource for a given type. This also satisfies the discovery.Source
// interface
type SourceFunction func(cs *kubernetes.Clientset) (ResourceSource, error)

// SourceFunctions is the list of functions to load
var SourceFunctions = []SourceFunction{
	PodSource,
	ServiceSource,
	PVCSource,
	SecretSource,
	EndpointSource,
	ServiceAccountSource,
	LimitRangeSource,
	ReplicationControllerSource,
	ResourceQuotaSource,
	DaemonSetSource,
	ReplicaSetSource,
	DeploymentSource,
	HorizontalPodAutoscalerSource,
	StatefulSetSource,
	IngressSource,
	NetworkPolicySource,
	PodDisruptionBudgetSource,
	RoleBindingSource,
	RoleSource,
	EndpointSliceSource,
	NamespaceSource,
	NodeSource,
	PersistentVolumeSource,
	StorageClassSource,
	PriorityClassSource,
}

// ResourceSource represents a source of Kubernetes resources. one of these
// sources needs to be created, and then have its get and list functions
// registered by calling the LoadFunctions method. Note that in order for this
// to be able to discover any time of Kubernetes resource it uses as significant
// amount of reflection. The LoadFunctions method should do enough error
// checking to ensure that the methods on this struct don't cause any panics,
// but there is still a very real chance that there will be panics so be
// careful doing anything non-standard with this struct
type ResourceSource struct {
	// The type of items that will be returned from this source
	ItemType string
	// A function that will accept an interface and return a list of items. The
	// interface that is passed will be the first item returned from
	// "listFunction", as an interface. The function should covert this to
	// whatever format it is expecting, the proceed to map to Items
	MapList func(interface{}) ([]*sdp.Item, error)
	// A function that will accept an interface and return an item. The
	// interface that is passed will be the first item returned from
	// "getFunction", as an interface. The function should covert this to
	// whatever format it is expecting, the proceed to map to an item
	MapGet func(interface{}) (*sdp.Item, error)

	// Whether or not this source is for namespaced resources
	Namespaced bool

	// NSS Namespace storage for when backends need to lookup the list of
	// namespaces
	NSS *NamespaceStorage

	// interfaceFunction is used to store the function which, when called,
	// returns an interface that we can call Get() and List() against in order
	// to get item details
	interfaceFunction reflect.Value
}

// LoadFunctions performs validation on the supplied interface function. This
// function should retrun an interface which has Get() and List() methods
//
// A Get should be:
//
//	func(ctx context.Context, name string, opts metaV1.GetOptions)
//
// A List should be:
//
//	List(ctx context.Context, opts metaV1.ListOptions)
func (rs *ResourceSource) LoadFunction(interfaceFunction interface{}) error {
	// Reflect to values
	interfaceFunctionValue := reflect.ValueOf(interfaceFunction)
	interfaceFunctionType := reflect.TypeOf(interfaceFunction)

	switch interfaceFunctionType.NumIn() {
	case 0:
		// Do nothing
	case 1:
		if interfaceFunctionType.In(0).Kind() != reflect.String {
			return errors.New("interfaceFunction first argument must be a string")
		}
	default:
		return errors.New("interfaceFunction should have 0 or 1 parameters")
	}

	if interfaceFunctionType.Out(0).Kind() != reflect.Interface {
		return errors.New("interfaceFunction return value should be an interface")
	}

	// This is the value that is going ot be returned when the interface
	// function is called. We need to check that this has the methods that we
	// expect and is therefore going to work when we try to interact with it
	returnInterface := interfaceFunctionType.Out(0)

	getMethod, getFound := returnInterface.MethodByName("Get")

	if !getFound {
		return errors.New("interfaceFunction does not have a 'Get' method")
	}

	listMethod, listFound := returnInterface.MethodByName("List")

	if !listFound {
		return errors.New("interfaceFunction does not have a 'List' method")
	}

	getFunctionType := getMethod.Type
	listFunctionType := listMethod.Type

	if getFunctionType.NumIn() != 3 {
		return errors.New("getFunction must accept 3 arguments")
	}

	// Check that parameters are as expected
	if getFunctionType.In(0).Kind() != reflect.Interface {
		return errors.New("getFunction first argument must be a context.Context")
	}

	if getFunctionType.In(1).Kind() != reflect.String {
		return errors.New("getFunction second argument must be a string")
	}

	if getFunctionType.In(2).Kind() != reflect.Struct {
		return errors.New("getFunction third argument must be a metaV1.GetOptions")
	}

	if getFunctionType.NumOut() != 2 {
		return errors.New("getFunction must return 2 values")
	}

	if listFunctionType.NumIn() != 2 {
		return errors.New("listFunction must accept 2 arguments")
	}

	if listFunctionType.In(0).Kind() != reflect.Interface {
		return errors.New("listFunction first argument must be a context.Context")
	}

	if listFunctionType.In(1).Kind() != reflect.Struct {
		return errors.New("getFunction second argument must be a metaV1.ListOptions")
	}

	if listFunctionType.NumOut() != 2 {
		return errors.New("listFunction must return 2 values")
	}

	// Save values for later use
	rs.interfaceFunction = interfaceFunctionValue

	return nil
}

// Get takes the UniqueAttribute value as a parameter (also referred to as the
// "name" of the item) and returns a full item will all details. This function
// must return an item whose UniqueAttribute value exactly matches the supplied
// parameter. If the item cannot be found it should return an ItemNotFoundError
// (Required)
func (rs *ResourceSource) Get(ctx context.Context, itemScope string, name string) (*sdp.Item, error) {
	var ctxValue reflect.Value
	var opts metaV1.GetOptions
	var optsValue reflect.Value
	var nameValue reflect.Value
	var params []reflect.Value
	var returns []reflect.Value
	var function reflect.Value
	var err error

	opts = metaV1.GetOptions{}

	// TODO: Logging

	ctxValue = reflect.ValueOf(ctx)
	nameValue = reflect.ValueOf(name)
	optsValue = reflect.ValueOf(opts)
	params = []reflect.Value{
		ctxValue,
		nameValue,
		optsValue,
	}

	// Call the function
	function, err = rs.getFunction(itemScope)

	if err != nil {
		return nil, err
	}

	returns = function.Call(params)

	if e := returns[1].Interface(); e != nil {
		if err, ok := e.(error); ok {
			return &sdp.Item{}, err
		}
		return &sdp.Item{}, errors.New("unknown error occurred")
	}

	// Map results and return
	return rs.MapGet(returns[0].Interface())
}

// List finds all items that the backend possibly can. It maybe be possible that
// this might not be an exhaustive list though in the case of kubernetes it is
// unlikely
func (rs *ResourceSource) List(ctx context.Context, itemScope string) ([]*sdp.Item, error) {
	var ctxValue reflect.Value
	var opts metaV1.ListOptions
	var optsValue reflect.Value
	var params []reflect.Value
	var function reflect.Value
	var returns []reflect.Value
	var err error

	opts = metaV1.ListOptions{}

	// TODO: Logging

	ctxValue = reflect.ValueOf(ctx)
	optsValue = reflect.ValueOf(opts)
	params = []reflect.Value{
		ctxValue,
		optsValue,
	}

	// TODO: The below relies on being able to parse out the scope from the
	// query. However it's entirely possible that the scope could be '*', so
	// we need to be able to handle that

	// Call the function
	function, err = rs.listFunction(itemScope)

	if err != nil {
		return nil, err
	}

	returns = function.Call(params)

	// Check if the error is nil. If it's nil then we know there wasn't an
	// error. If not then we know there was an error
	if returns[1].Interface() != nil {
		return make([]*sdp.Item, 0), returns[1].Interface().(error)
	}

	return rs.MapList(returns[0].Interface())
}

// Search This search for items that match a given ListOptions. The query must
// be provided as a JSON object that can be cast to a
// [ListOptions](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#ListOptions)
// object.
//
// *Note:* Additional changes will be made to the ListOptions object after
// deserialization such as limiting the scope to items of the same type as the
// current ResourceSource, and drooping any options such as "Watch"
func (rs *ResourceSource) Search(ctx context.Context, itemScope string, query string) ([]*sdp.Item, error) {
	var ctxValue reflect.Value
	var opts metaV1.ListOptions
	var optsValue reflect.Value
	var params []reflect.Value
	var returns []reflect.Value
	var function reflect.Value
	var err error

	opts, err = QueryToListOptions(query)

	if err != nil {
		log.WithFields(log.Fields{
			"query":      query,
			"type":       rs.ItemType,
			"scope":      itemScope,
			"parseError": err.Error(),
		}).Error("error while parsing query")

		return nil, err
	}

	// TODO: Logging

	ctxValue = reflect.ValueOf(ctx)
	optsValue = reflect.ValueOf(opts)
	params = []reflect.Value{
		ctxValue,
		optsValue,
	}

	// Call the function
	function, err = rs.listFunction(itemScope)

	if err != nil {
		return nil, err
	}

	returns = function.Call(params)

	// Check for an error
	if returns[1].Interface() != nil {
		return make([]*sdp.Item, 0), returns[1].Interface().(error)
	}

	return rs.MapList(returns[0].Interface())
}

// Type is the type of items that this returns (Required)
func (rs *ResourceSource) Type() string {
	return rs.ItemType
}

// Name returns a descriptive name for the source, used in logging and metadata
func (rs *ResourceSource) Name() string {
	return fmt.Sprintf("k8s-%v", rs.ItemType)
}

// Scopes Returns the list of scops that this source is capable of finding
// items for. This is usually the name of the cluster, plus any namespaces in
// the format {clusterName}.{namespace}
func (rs *ResourceSource) Scopes() []string {
	contexts := make([]string, 0)

	if rs.Namespaced {
		namespaces, _ := rs.NSS.Namespaces()

		for _, namespace := range namespaces {
			contexts = append(contexts, ClusterName+"."+namespace)
		}
	} else {
		contexts = append(contexts, ClusterName)
	}

	return contexts
}

// Weight The weight of this source, used for conflict resolution. Currently
// returns a static value of 100
func (rs *ResourceSource) Weight() int {
	return 100
}

// interactionInterface Calls the interface function to return an interface that
// will allow us to call Get and List functions which will in turn actually
// execute API queries against K8s
func (rs *ResourceSource) interactionInterface(itemScope string) (reflect.Value, error) {
	contextDetails, _ := ParseScope(itemScope, rs.Namespaced)
	interfaceFunctionArgs := make([]reflect.Value, 0)

	if rs.Namespaced {
		// If the interface function is namespaced we need to pass in the namespace that we want to query
		interfaceFunctionArgs = append(interfaceFunctionArgs, reflect.ValueOf(contextDetails.Namespace))
	}

	// Call the interface function in order to return a list function for the
	// given namespace (or not, if the source isn't namespaced)
	results := rs.interfaceFunction.Call(interfaceFunctionArgs)

	// Validate the results before sending them back
	if len(results) != 1 {
		return reflect.Value{}, errors.New("could not load list function, loading returned too many results")
	}

	return results[0], nil
}

func (rs *ResourceSource) getFunction(itemScope string) (reflect.Value, error) {
	var getMethod reflect.Value
	var iFace reflect.Value
	var err error

	iFace, err = rs.interactionInterface(itemScope)

	if err != nil {
		return reflect.Value{}, err
	}

	getMethod = iFace.MethodByName("Get")

	return getMethod, nil
}

func (rs *ResourceSource) listFunction(itemScope string) (reflect.Value, error) {
	var listMethod reflect.Value
	var iFace reflect.Value
	var err error

	iFace, err = rs.interactionInterface(itemScope)

	if err != nil {
		return reflect.Value{}, err
	}

	listMethod = iFace.MethodByName("List")

	return listMethod, nil
}
