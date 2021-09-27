package backends

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/dylanratcliffe/source-go/pkg/sources"
)

// APITimeoutContext Returns a context representing the configured timeout for
// API calls. This is configured in the "apitimeout" setting and is represented
// in seconds
func APITimeoutContext() (context.Context, context.CancelFunc) {
	if apiTimeoutSet {
		return context.WithTimeout(context.Background(), apiTimeout)
	}

	// If the config hasn't been set yet then load it
	apiTimeout = apiTimeoutDefault

	if t := sources.ConfigGetInt("apitimeout", BackendPackage); t != 0 {
		// If a timeout has been set then use that
		apiTimeout = time.Duration(t) * time.Second
	}

	// Note that we have set it so we don't have to go through this again
	apiTimeoutSet = true

	return APITimeoutContext()
}

// GetAllNamespaceNames gets the names of all namespaces that can be found for a
// given connection
func GetAllNamespaceNames(cs *kubernetes.Clientset) ([]string, error) {
	var opts metaV1.ListOptions
	var namespaces []string
	var namespaceList *coreV1.NamespaceList
	var err error

	// Use the configured API timeout
	ctx, cancel := APITimeoutContext()
	defer cancel()

	log.WithFields(log.Fields{
		"APIVersion": cs.CoreV1().RESTClient().APIVersion().Version,
		"APIURL":     cs.CoreV1().RESTClient().Get().URL(),
	}).Trace("Getting list of namepaces")

	// Get the list of namespaces
	namespaceList, err = cs.CoreV1().Namespaces().List(ctx, opts)

	if err != nil {
		return []string{}, err
	}

	for _, item := range namespaceList.Items {
		namespaces = append(namespaces, item.Name)
	}

	return namespaces, nil
}
