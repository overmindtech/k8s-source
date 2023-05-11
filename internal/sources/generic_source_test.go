package sources

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
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
		},
	}, nil
}

func createSource() KubeTypeSource[*v1.Pod, *v1.PodList] {
	return KubeTypeSource[*v1.Pod, *v1.PodList]{
		ClusterInterfaceBuilder: func() ItemInterface[*v1.Pod, *v1.PodList] {
			return PodClient{}
		},
		ListExtractor: func(p *v1.PodList) ([]*v1.Pod, error) {
			pods := make([]*v1.Pod, len(p.Items))

			for i := range p.Items {
				pods[i] = &p.Items[i]
			}

			return pods, nil
		},
		LinkedItemQueryExtractor: func(p *v1.Pod) ([]*sdp.Query, error) {
			queries := make([]*sdp.Query, 0)

			if p.Spec.NodeName == "" {
				queries = append(queries, &sdp.Query{
					Type:   "node",
					Method: sdp.QueryMethod_GET,
					Query:  p.Spec.NodeName,
					Scope:  "foo",
				})
			}

			return queries, nil
		},
		TypeName:    "pod",
		ClusterName: "minikube",
	}
}

func TestSourceValidate(t *testing.T) {
	t.Run("fully populated source", func(t *testing.T) {
		t.Parallel()
		source := createSource()
		err := source.Validate()

		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}
	})

	t.Run("missing ClusterInterfaceBuilder", func(t *testing.T) {
		t.Parallel()
		source := createSource()
		source.ClusterInterfaceBuilder = nil

		err := source.Validate()

		if err == nil {
			t.Errorf("expected error, got none")
		}
	})

	t.Run("missing ListExtractor", func(t *testing.T) {
		t.Parallel()
		source := createSource()
		source.ListExtractor = nil

		err := source.Validate()

		if err == nil {
			t.Errorf("expected error, got none")
		}
	})

	t.Run("missing TypeName", func(t *testing.T) {
		t.Parallel()
		source := createSource()
		source.TypeName = ""

		err := source.Validate()

		if err == nil {
			t.Errorf("expected error, got none")
		}
	})
}

func TestSourceGet(t *testing.T) {
	t.Run("get existing item", func(t *testing.T) {
		source := createSource()

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
	})

	t.Run("get non-existent item", func(t *testing.T) {
		source := createSource()
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

func TestRealGet(t *testing.T) {
	source := KubeTypeSource[*v1.Pod, *v1.PodList]{
		TypeName:    "pod",
		Namespaces:  []string{"default"},
		ClusterName: "minikube",
		NamespacedInterfaceBuilder: func(namespace string) ItemInterface[*v1.Pod, *v1.PodList] {
			return CurrentCluster.ClientSet.CoreV1().Pods(namespace)
		},
		ListExtractor: func(p *v1.PodList) ([]*v1.Pod, error) {
			pods := make([]*v1.Pod, len(p.Items))

			for i := range p.Items {
				pods[i] = &p.Items[i]
			}

			return pods, nil
		},
		LinkedItemQueryExtractor: func(p *v1.Pod) ([]*sdp.Query, error) {
			return []*sdp.Query{}, nil
		},
	}

	err := source.Validate()

	if err != nil {
		t.Fatalf("source validation failed: %s", err)
	}

	_, err = source.Get(context.Background(), "minikube:8080.default", "not-real-pod")

	if err == nil {
		t.Error("expected error, got none")
	}

	sdpErr := new(sdp.QueryError)

	if errors.As(err, &sdpErr) {
		if sdpErr.ErrorType != sdp.QueryError_NOTFOUND {
			t.Errorf("expected not found error, got %s", sdpErr.ErrorType)
		}
	} else {
		t.Errorf("expected sdp.QueryError, got %s", err)
	}
}
