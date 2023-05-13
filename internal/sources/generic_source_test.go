package sources

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type PodClient struct {
	GetError  error
	ListError error
}

func (p PodClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Pod, error) {
	if p.GetError != nil {
		return nil, p.GetError
	}

	uid := uuid.NewString()

	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			UID:               types.UID(uid),
			ResourceVersion:   "9164",
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: "kube-api-access-hgq4d",
				},
			},
			RestartPolicy:      "Always",
			DNSPolicy:          "ClusterFirst",
			ServiceAccountName: "default",
			NodeName:           "minikube",
		},
		Status: v1.PodStatus{
			Phase:  "Running",
			HostIP: "10.0.0.1",
			PodIP:  "10.244.0.6",
		},
	}, nil
}

func (p PodClient) List(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error) {
	if p.ListError != nil {
		return nil, p.ListError
	}

	uid := uuid.NewString()

	return &v1.PodList{
		Items: []v1.Pod{
			{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "foo",
					Namespace:         "default",
					UID:               types.UID(uid),
					ResourceVersion:   "9164",
					CreationTimestamp: metav1.NewTime(time.Now()),
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "kube-api-access-hgq4d",
						},
					},
					RestartPolicy:      "Always",
					DNSPolicy:          "ClusterFirst",
					ServiceAccountName: "default",
					NodeName:           "minikube",
				},
				Status: v1.PodStatus{
					Phase:  "Running",
					HostIP: "10.0.0.1",
					PodIP:  "10.244.0.6",
				},
			},
			{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "bar",
					Namespace:         "default",
					UID:               types.UID(uid),
					ResourceVersion:   "9164",
					CreationTimestamp: metav1.NewTime(time.Now()),
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "kube-api-access-c43w1",
						},
					},
					RestartPolicy:      "Always",
					DNSPolicy:          "ClusterFirst",
					ServiceAccountName: "default",
					NodeName:           "minikube",
				},
				Status: v1.PodStatus{
					Phase:  "Running",
					HostIP: "10.0.0.1",
					PodIP:  "10.244.0.7",
				},
			},
		},
	}, nil
}

func createSource(namespaced bool) KubeTypeSource[*v1.Pod, *v1.PodList] {
	var clusterInterfaceBuilder ClusterInterfaceBuilder[*v1.Pod, *v1.PodList]
	var namespacedInterfaceBuilder NamespacedInterfaceBuilder[*v1.Pod, *v1.PodList]

	if namespaced {
		namespacedInterfaceBuilder = func(namespace string) ItemInterface[*v1.Pod, *v1.PodList] {
			return PodClient{}
		}
	} else {
		clusterInterfaceBuilder = func() ItemInterface[*v1.Pod, *v1.PodList] {
			return PodClient{}
		}
	}

	return KubeTypeSource[*v1.Pod, *v1.PodList]{
		ClusterInterfaceBuilder:    clusterInterfaceBuilder,
		NamespacedInterfaceBuilder: namespacedInterfaceBuilder,
		ListExtractor: func(p *v1.PodList) ([]*v1.Pod, error) {
			pods := make([]*v1.Pod, len(p.Items))

			for i := range p.Items {
				pods[i] = &p.Items[i]
			}

			return pods, nil
		},
		LinkedItemQueryExtractor: func(p *v1.Pod, scope string) ([]*sdp.Query, error) {
			queries := make([]*sdp.Query, 0)

			if p.Spec.NodeName == "" {
				queries = append(queries, &sdp.Query{
					Type:   "node",
					Method: sdp.QueryMethod_GET,
					Query:  p.Spec.NodeName,
					Scope:  scope,
				})
			}

			return queries, nil
		},
		HealthExtractor: func(resource *v1.Pod) *sdp.Health {
			return sdp.Health_HEALTH_OK.Enum()
		},
		TypeName:    "Pod",
		ClusterName: "minikube",
		Namespaces:  []string{"default", "app1"},
	}
}

func TestSourceValidate(t *testing.T) {
	t.Run("fully populated source", func(t *testing.T) {
		t.Parallel()
		source := createSource(false)
		err := source.Validate()

		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}
	})

	t.Run("missing ClusterInterfaceBuilder", func(t *testing.T) {
		t.Parallel()
		source := createSource(false)
		source.ClusterInterfaceBuilder = nil

		err := source.Validate()

		if err == nil {
			t.Errorf("expected error, got none")
		}
	})

	t.Run("missing ListExtractor", func(t *testing.T) {
		t.Parallel()
		source := createSource(false)
		source.ListExtractor = nil

		err := source.Validate()

		if err == nil {
			t.Errorf("expected error, got none")
		}
	})

	t.Run("missing TypeName", func(t *testing.T) {
		t.Parallel()
		source := createSource(false)
		source.TypeName = ""

		err := source.Validate()

		if err == nil {
			t.Errorf("expected error, got none")
		}
	})

	t.Run("missing ClusterName", func(t *testing.T) {
		t.Parallel()
		source := createSource(false)
		source.ClusterName = ""

		err := source.Validate()

		if err == nil {
			t.Errorf("expected error, got none")
		}
	})

	t.Run("missing namespaces", func(t *testing.T) {
		t.Run("when namespaced", func(t *testing.T) {
			t.Parallel()
			source := createSource(true)
			source.Namespaces = nil

			err := source.Validate()

			if err == nil {
				t.Errorf("expected error, got none")
			}

			source.Namespaces = []string{}

			err = source.Validate()

			if err == nil {
				t.Errorf("expected error, got none")
			}
		})

		t.Run("when not namespaced", func(t *testing.T) {
			t.Parallel()
			source := createSource(false)
			source.Namespaces = nil

			err := source.Validate()

			if err != nil {
				t.Errorf("expected no error, got %s", err)
			}

			source.Namespaces = []string{}

			err = source.Validate()

			if err != nil {
				t.Errorf("expected no error, got %s", err)
			}
		})

	})
}

func TestType(t *testing.T) {
	source := createSource(false)

	if source.Type() != "Pod" {
		t.Errorf("expected type 'Pod', got %s", source.Type())
	}
}

func TestName(t *testing.T) {
	source := createSource(false)

	if source.Name() == "" {
		t.Errorf("expected non-empty name, got none")
	}
}

func TestScopes(t *testing.T) {
	t.Run("when namespaced", func(t *testing.T) {
		source := createSource(true)

		if len(source.Scopes()) != len(source.Namespaces) {
			t.Errorf("expected %d scopes, got %d", len(source.Namespaces), len(source.Scopes()))
		}
	})

	t.Run("when not namespaced", func(t *testing.T) {
		source := createSource(false)

		if len(source.Scopes()) != 1 {
			t.Errorf("expected 1 scope, got %d", len(source.Scopes()))
		}
	})
}

func TestSourceGet(t *testing.T) {
	t.Run("get existing item", func(t *testing.T) {
		source := createSource(false)

		item, err := source.Get(context.Background(), "foo", "example")

		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if item == nil {
			t.Errorf("expected item, got none")
		}

		if item.UniqueAttributeValue() != "example" {
			t.Errorf("expected item with unique attribute value 'example', got %s", item.UniqueAttributeValue())
		}

		if *item.Health != sdp.Health_HEALTH_OK {
			t.Errorf("expected item with health HEALTH_OK, got %s", item.Health)
		}
	})

	t.Run("get non-existent item", func(t *testing.T) {
		source := createSource(false)
		source.ClusterInterfaceBuilder = func() ItemInterface[*v1.Pod, *v1.PodList] {
			return PodClient{
				GetError: &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: "not found",
				},
			}
		}

		_, err := source.Get(context.Background(), "foo", "example")

		if err == nil {
			t.Errorf("expected error, got none")
		}
	})
}

func TestFailingQueryExtractor(t *testing.T) {
	source := createSource(false)
	source.LinkedItemQueryExtractor = func(_ *v1.Pod, _ string) ([]*sdp.Query, error) {
		return nil, errors.New("failed to extract queries")
	}

	_, err := source.Get(context.Background(), "foo", "example")

	if err == nil {
		t.Errorf("expected error, got none")
	}
}

func TestList(t *testing.T) {
	t.Run("when namespaced", func(t *testing.T) {
		source := createSource(true)

		items, err := source.List(context.Background(), "foo.bar")

		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}

		if *items[0].Health != sdp.Health_HEALTH_OK {
			t.Errorf("expected item with health HEALTH_OK, got %s", items[0].Health)
		}
	})

	t.Run("when not namespaced", func(t *testing.T) {
		source := createSource(false)

		items, err := source.List(context.Background(), "foo")

		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}
	})

	t.Run("with failing list extractor", func(t *testing.T) {
		source := createSource(false)
		source.ListExtractor = func(_ *v1.PodList) ([]*v1.Pod, error) {
			return nil, errors.New("failed to extract list")
		}

		_, err := source.List(context.Background(), "foo")

		if err == nil {
			t.Errorf("expected error, got none")
		}
	})

	t.Run("with failing query extractor", func(t *testing.T) {
		source := createSource(false)
		source.LinkedItemQueryExtractor = func(_ *v1.Pod, _ string) ([]*sdp.Query, error) {
			return nil, errors.New("failed to extract queries")
		}

		_, err := source.List(context.Background(), "foo")

		if err == nil {
			t.Errorf("expected error, got none")
		}
	})
}

func TestSearch(t *testing.T) {
	t.Run("with a valid query", func(t *testing.T) {
		source := createSource(false)

		items, err := source.Search(context.Background(), "foo", "{\"labelSelector\":\"app=foo\"}")

		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		if len(items) != 2 {
			t.Errorf("expected 2 item, got %d", len(items))
		}
	})

	t.Run("with an invalid query", func(t *testing.T) {
		source := createSource(false)

		_, err := source.Search(context.Background(), "foo", "{{{{}")

		if err == nil {
			t.Errorf("expected error, got none")
		}
	})
}

func TestRedact(t *testing.T) {
	source := createSource(true)
	source.Redact = func(resource *v1.Pod) *v1.Pod {
		resource.Spec.Hostname = "redacted"

		return resource
	}

	item, err := source.Get(context.Background(), "cluster.namespace", "test")

	if err != nil {
		t.Error(err)
	}

	hostname, err := item.Attributes.Get("spec.hostname")

	if err != nil {
		t.Error(err)
	}

	if hostname != "redacted" {
		t.Errorf("expected hostname to be redacted, got %v", hostname)
	}
}

type QueryTest struct {
	ExpectedType   string
	ExpectedMethod sdp.QueryMethod
	ExpectedQuery  string
	ExpectedScope  string

	// Expect the query to match a regex, this takes precedence over
	// ExpectedQuery
	ExpectedQueryMatches *regexp.Regexp
}

type QueryTests []QueryTest

func (i QueryTests) Execute(t *testing.T, item *sdp.Item) {
	for _, test := range i {
		var found bool

		for _, lir := range item.LinkedItemQueries {
			if lirMatches(test, lir) {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("could not find linked item request in %v requests.\nType: %v\nQuery: %v\nScope: %v", len(item.LinkedItemQueries), test.ExpectedType, test.ExpectedQuery, test.ExpectedScope)
		}
	}
}

func lirMatches(test QueryTest, req *sdp.Query) bool {
	methodOK := test.ExpectedMethod == req.Method
	scopeOK := test.ExpectedScope == req.Scope
	typeOK := test.ExpectedType == req.Type
	var queryOK bool

	if test.ExpectedQueryMatches != nil {
		queryOK = test.ExpectedQueryMatches.MatchString(req.Query)
	} else {
		queryOK = test.ExpectedQuery == req.Query
	}

	return methodOK && scopeOK && typeOK && queryOK
}

type SourceTests struct {
	// The source under test
	Source discovery.Source

	// The get query to test
	GetQuery      string
	GetScope      string
	GetQueryTests QueryTests

	// If this is set,. the get query is determined by running a list, then
	// finding the first item that matches this regexp
	GetQueryRegexp *regexp.Regexp

	// YAML to apply before testing, it will be removed after
	SetupYAML string
}

func (s SourceTests) Execute(t *testing.T) {
	t.Parallel()

	if s.SetupYAML != "" {
		err := CurrentCluster.Apply(s.SetupYAML)

		if err != nil {
			t.Fatal(err)
		}

		t.Cleanup(func() {
			CurrentCluster.Delete(s.SetupYAML)
		})
	}

	t.Run(s.Source.Name(), func(t *testing.T) {
		var getQuery string

		if s.GetQueryRegexp != nil {
			items, err := s.Source.List(context.Background(), s.GetScope)

			if err != nil {
				t.Fatal(err)
			}

			for _, item := range items {
				if s.GetQueryRegexp.MatchString(item.UniqueAttributeValue()) {
					getQuery = item.UniqueAttributeValue()
					break
				}
			}
		} else {
			getQuery = s.GetQuery
		}

		if getQuery != "" {
			t.Run(fmt.Sprintf("GET:%v", getQuery), func(t *testing.T) {
				item, err := s.Source.Get(context.Background(), s.GetScope, getQuery)

				if err != nil {
					t.Fatal(err)
				}

				if item == nil {
					t.Errorf("expected item, got none")
				}

				if err = item.Validate(); err != nil {
					t.Error(err)
				}

				s.GetQueryTests.Execute(t, item)
			})
		}

		t.Run("LIST", func(t *testing.T) {
			items, err := s.Source.List(context.Background(), s.GetScope)

			if err != nil {
				t.Fatal(err)
			}

			if len(items) == 0 {
				t.Errorf("expected items, got none")
			}

			itemMap := make(map[string]*sdp.Item)

			for _, item := range items {
				itemMap[item.UniqueAttributeValue()] = item

				if err = item.Validate(); err != nil {
					t.Error(err)
				}
			}

			if len(itemMap) != len(items) {
				t.Errorf("expected %v unique items, got %v", len(items), len(itemMap))
			}
		})
	})
}

// WaitFor waits for a condition to be true, or returns an error if the timeout
func WaitFor(timeout time.Duration, run func() bool) error {
	start := time.Now()

	for {
		if run() {
			return nil
		}

		if time.Since(start) > timeout {
			return fmt.Errorf("timeout exceeded")
		}

		time.Sleep(250 * time.Millisecond)
	}
}

func TestObjectReferenceToQuery(t *testing.T) {
	t.Run("with a valid object reference", func(t *testing.T) {
		ref := &v1.ObjectReference{
			Kind:      "Pod",
			Namespace: "default",
			Name:      "foo",
		}

		query := ObjectReferenceToQuery(ref, ScopeDetails{
			ClusterName: "test-cluster",
			Namespace:   "default",
		})

		if query.Type != "Pod" {
			t.Errorf("expected type Pod, got %s", query.Type)
		}

		if query.Query != "foo" {
			t.Errorf("expected query to be foo, got %s", query.Query)
		}

		if query.Scope != "test-cluster.default" {
			t.Errorf("expected scope to be test-cluster.default, got %s", query.Scope)
		}
	})

	t.Run("with a nil object reference", func(t *testing.T) {
		query := ObjectReferenceToQuery(nil, ScopeDetails{})

		if query != nil {
			t.Errorf("expected nil query, got %v", query)
		}
	})
}
