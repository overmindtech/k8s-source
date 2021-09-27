package backends

import (
	"sync"
	"time"

	"k8s.io/client-go/kubernetes"
)

// NamespaceStorage is an object that is used for storing the list of
// namespaces. Many of the non-namespaced resources will need to know what thee
// list of namespaces is in order for them to create LinkedItemRequests. This
// object stores them to ensure that we don't unnecessarily spam the API
type NamespaceStorage struct {
	CS            *kubernetes.Clientset
	CacheDuration time.Duration

	lastUpdate      time.Time
	namespaces      []string
	namespacesMutex sync.RWMutex
}

// Namespaces returns the list of namespaces, updating if required
func (ns *NamespaceStorage) Namespaces() ([]string, error) {
	ns.namespacesMutex.RLock()

	// Check that the cache is up to date
	if time.Since(ns.lastUpdate) < ns.CacheDuration {
		defer ns.namespacesMutex.RUnlock()
		return ns.namespaces, nil
	}

	ns.namespacesMutex.RUnlock()
	ns.namespacesMutex.Lock()

	var err error

	// Call the API
	ns.namespaces, err = GetAllNamespaceNames(ns.CS)

	if err != nil {
		ns.namespacesMutex.Unlock()
		return ns.namespaces, err
	}

	// Update dates
	ns.lastUpdate = time.Now()
	ns.namespacesMutex.Unlock()

	return ns.Namespaces()
}
