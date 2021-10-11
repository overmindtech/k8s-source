package sources

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/dylanratcliffe/sdp-go"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	log "github.com/sirupsen/logrus"
)

// defaultAPITimeout is the default amount of time to wait per for each API
// query to K8s. This is passed as a context to API request functions
var apiTimeoutDefault = (10 * time.Second)
var apiTimeoutSet = false
var apiTimeout time.Duration

// BackendPackage is the name of this backend package
const BackendPackage = "k8s"

// ClusterName stores the name of the cluster, this is also used as the context
// for non-namespaced items. This designed to be user by namespaced items to
// create linked item requests on non-namespaced items
var ClusterName string

// NamespacedSourceFunction is a function that accepts a kubernetes client and
// namespace, and returns a ResourceSource for a given type. This also satisfies
// the Backend interface
type NamespacedSourceFunction func(cs *kubernetes.Clientset, namespace string) ResourceSource

// NonNamespacedSourceFunction is a function that accepts a kubernetes client and
// returns a ResourceSource for a given type. This also satisfies the Backend
// interface
type NonNamespacedSourceFunction func(cs *kubernetes.Clientset, nss *NamespaceStorage) ResourceSource

// NamespacedSourceFunctions is the list of functions to load
var NamespacedSourceFunctions = []NamespacedSourceFunction{
	PodSource,
	ServiceSource,
	PVCSource,
	SecretSource,
	ConfigMapSource,
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
	JobSource,
	CronJobSource,
	IngressSource,
	NetworkPolicySource,
	PodDisruptionBudgetSource,
	RoleBindingSource,
	RoleSource,
	EndpointSliceSource,
}

// NonNamespacedSourceFunctions is the list of functions to load
var NonNamespacedSourceFunctions = []NonNamespacedSourceFunction{
	NamespaceSource,
	NodeSource,
	PersistentVolumeSource,
	ClusterRoleSource,
	ClusterRoleBindingSource,
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
	// the context of the items that will be returned by this source
	ItemContext string
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

	// NSS Namespace storage for when backends need to lookup the list of
	// namespaces
	NSS *NamespaceStorage

	// getFunction and listFunction are populated by LoadFunctions
	getFunction  reflect.Value
	listFunction reflect.Value
}

// LoadFunctions performs validation on the supplied Get and List functions,
// ensuring that they have valid inputs and outputs.
//
// A get should be:
//   func(ctx context.Context, name string, opts metaV1.GetOptions)
//
// A List should be:
//   List(ctx context.Context, opts metaV1.ListOptions)
//
func (rs *ResourceSource) LoadFunctions(getFunction interface{}, listFunction interface{}) error {
	// Reflect to values
	getFunctionValue := reflect.ValueOf(getFunction)
	listFunctionValue := reflect.ValueOf(listFunction)
	getFunctionType := reflect.TypeOf(getFunction)
	listFunctionType := reflect.TypeOf(listFunction)

	// Validate that they are functions
	if getFunctionValue.Kind() != reflect.Func {
		return errors.New("getFunction is not a Func")
	}

	if listFunctionValue.Kind() != reflect.Func {
		return errors.New("listFunction is not a Func")
	}

	if getFunctionType.NumIn() != 3 {
		return errors.New("getFunction must accept 3 arguments")
	}

	// Check that paramaters are as expected
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
	rs.getFunction = getFunctionValue
	rs.listFunction = listFunctionValue

	return nil
}

// Get takes the UniqueAttribute value as a parameter (also referred to as the
// "name" of the item) and returns a full item will all details. This function
// must return an item whose UniqueAttribute value exactly matches the supplied
// parameter. If the item cannot be found it should return an ItemNotFoundError
// (Required)
func (rs *ResourceSource) Get(itemContext string, name string) (*sdp.Item, error) {
	var ctx context.Context
	var ctxValue reflect.Value
	var opts metaV1.GetOptions
	var optsValue reflect.Value
	var nameValue reflect.Value
	var params []reflect.Value
	var returns []reflect.Value

	// TODO: Add API timeout
	ctx = context.Background()
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
	returns = rs.getFunction.Call(params)

	if e := returns[1].Interface(); e != nil {
		if err, ok := e.(error); ok {
			return &sdp.Item{}, err
		}
		return &sdp.Item{}, errors.New("unknown error occurred")
	}

	// Map results and return
	return rs.MapGet(returns[0].Interface())
}

// Find finds all items that the backend possibly can. It maybe be possible that
// this might not be an exhaustive list though in the case of kubernetes it is
// unlikely
func (rs *ResourceSource) Find() ([]*sdp.Item, error) {
	var ctx context.Context
	var ctxValue reflect.Value
	var opts metaV1.ListOptions
	var optsValue reflect.Value
	var params []reflect.Value
	var returns []reflect.Value

	// TODO: Add API timeout
	ctx = context.Background()
	opts = metaV1.ListOptions{}

	// TODO: Logging

	ctxValue = reflect.ValueOf(ctx)
	optsValue = reflect.ValueOf(opts)
	params = []reflect.Value{
		ctxValue,
		optsValue,
	}

	// Call the function
	returns = rs.listFunction.Call(params)

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
func (rs *ResourceSource) Search(query string) ([]*sdp.Item, error) {
	var ctx context.Context
	var ctxValue reflect.Value
	var opts metaV1.ListOptions
	var optsValue reflect.Value
	var params []reflect.Value
	var returns []reflect.Value
	var err error

	// TODO: Add API timeout
	ctx = context.Background()
	opts, err = QueryToListOptions(query)

	if err != nil {
		log.WithFields(log.Fields{
			"query":      query,
			"type":       rs.ItemType,
			"context":    rs.ItemContext,
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
	returns = rs.listFunction.Call(params)

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

// BackendPackage Returns the name of the backend package. This is used for
// debugging and logging (Required)
func (rs *ResourceSource) BackendPackage() string {
	return BackendPackage
}

// Context If this backend is returning items for a context other than the
// "local" context (this usually means the machine that the backend is running
// on) then it should return what context it *is* for using this method. This
// will be useful for things like pulling data from many kubernetes clusters or
// namespaces, in this case the context would need to include this information
// to ensure that the items are always unique (Optional)
func (rs *ResourceSource) Context() string {
	return rs.ItemContext
}

// Backends is the main loader function for this backend package. It will be
// called when the package is loaded and will return all backends that this
// package provides. If a connection can't be made to kubernetes it simply won't
// return anything
// func Backends() ([]sources.Backend, error) {
// 	var err error
// 	var backends []sources.Backend
// 	var rc *rest.Config
// 	var clientSet *kubernetes.Clientset

// 	//
// 	// Connect to Kubernetes
// 	//

// 	// Load kube location from config
// 	kubeConfigPath := sources.ConfigGetString("kubeconfig", BackendPackage)

// 	// Check that we actually got something back and if not default to ~/.kube/config
// 	if kubeConfigPath == "" {
// 		home, err := homedir.Dir()

// 		if err != nil {
// 			return backends, err
// 		}

// 		kubeConfigPath = home + "/.kube/config"

// 	}

// 	// Load kubernetes config
// 	rc, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)

// 	if err != nil {
// 		return backends, err
// 	}

// 	// Create clientset
// 	clientSet, err = kubernetes.NewForConfig(rc)

// 	if err != nil {
// 		return backends, err
// 	}

// 	//
// 	// Discover info
// 	//
// 	// Now that we have a connection to the kubernetes cluster we need to go
// 	// about generating some backends for each context. In the case of
// 	// kubernetes the most obvious thing that we would divide contexts on is
// 	// namespace. However there are certain things that aren't namespaced, like
// 	// persistentVolmes, nodes etc. So these will need to have a context that
// 	// doesn't include the namespace
// 	var k8sURL *url.URL
// 	var k8sHost string
// 	var k8sPort string
// 	var nss NamespaceStorage
// 	var namespaces []string

// 	k8sURL, err = url.Parse(rc.Host)

// 	if err != nil {
// 		return []sources.Backend{}, err
// 	}

// 	// Calculate the cluster name
// 	k8sHost, k8sPort, err = net.SplitHostPort(k8sURL.Host)

// 	if err != nil {
// 		return nil, err
// 	}

// 	if k8sPort == "" || k8sPort == "443" {
// 		// If a port isn't specific or it's a strandard port then just return
// 		// the hostname
// 		ClusterName = k8sHost
// 	} else {
// 		// If it is running on a custom port then return host:port
// 		ClusterName = k8sHost + ":" + k8sPort
// 	}

// 	// Get list of namspaces
// 	nss = NamespaceStorage{
// 		CS:            clientSet,
// 		CacheDuration: (10 * time.Second),
// 	}

// 	namespaces, err = nss.Namespaces()

// 	if err != nil {
// 		// If we can't get namespaces then raise an error but keep going since
// 		// we might be able to get non-namespaced components
// 		log.WithFields(log.Fields{
// 			"underlying": err.Error(),
// 		}).Error("Failed to get namespaces, continuing")
// 	}

// 	// Load all non-namespaced backends
// 	for _, f := range NonNamespacedSourceFunctions {
// 		source := f(clientSet, &nss)

// 		source.ItemContext = ClusterName

// 		backends = append(backends, &source)
// 	}

// 	// Now that I have all of the namespaces I should be able to generate
// 	// backends for each type that is available.
// 	//
// 	// Firstly I need to range over the namespaces
// 	for _, namespace := range namespaces {
// 		context := ClusterName + "." + namespace

// 		for _, f := range NamespacedSourceFunctions {
// 			// Generate the source
// 			source := f(clientSet, namespace)

// 			// Assign context
// 			source.ItemContext = context

// 			backends = append(backends, &source)
// 		}
// 	}

// 	return backends, nil
// }
