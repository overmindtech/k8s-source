package sources

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
)

const TestNamespace = "k8s-source-testing"

const TestNamespaceYAML = `
apiVersion: v1
kind: Namespace
metadata:
  name: k8s-source-testing
`

type TestCluster struct {
	Name       string
	Kubeconfig string
	ClientSet  *kubernetes.Clientset
	provider   *cluster.Provider
	T          *testing.T
}

func (t *TestCluster) ConnectExisting(name string) error {
	kubeconfig := homedir.HomeDir() + "/.kube/config"

	var rc *rest.Config
	var err error

	// Load kubernetes config
	rc, err = clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		return err
	}

	var clientSet *kubernetes.Clientset

	// Create clientset
	clientSet, err = kubernetes.NewForConfig(rc)

	if err != nil {
		return err
	}

	// Validate that we can connect to the cluster
	_, err = clientSet.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})

	if err != nil {
		return err
	}

	t.Name = name
	t.Kubeconfig = kubeconfig
	t.ClientSet = clientSet

	return nil
}

func (t *TestCluster) Start() error {
	clusterName := "k8s-source-tests"

	log.Println("🔍 Trying to connect to existing cluster")
	err := t.ConnectExisting(clusterName)

	if err != nil {
		// If there is an error then create out own cluster
		log.Println("🤞 Creating Kubernetes cluster using Kind")

		t.provider = cluster.NewProvider()
		err = t.provider.Create(clusterName, cluster.CreateWithV1Alpha4Config(&v1alpha4.Cluster{}))

		if err != nil {
			return err
		}

		// Connect to the cluster we just created
		err = t.ConnectExisting(clusterName)

		if err != nil {
			return err
		}

		err = t.provider.ExportKubeConfig(t.Name, t.Kubeconfig, false)

		if err != nil {
			return err
		}
	}

	log.Printf("🐚 Ensuring test namespace %v exists\n", TestNamespace)
	err = t.Apply(TestNamespaceYAML)

	if err != nil {
		return err
	}

	return nil
}

func (t *TestCluster) ApplyBaselineConfig() error {
	return t.Apply(ClusterBaseline)
}

// Apply Runs of `kubectl apply -f` for a given string of YAML
func (t *TestCluster) Apply(yaml string) error {
	return t.kubectl("apply", yaml)
}

// Delete Runs of `kubectl delete -f` for a given string of YAML
func (t *TestCluster) Delete(yaml string) error {
	return t.kubectl("delete", yaml)
}

func (t *TestCluster) kubectl(method string, yaml string) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Create temp file to write config to
	config, err := os.CreateTemp("", "*-conf.yaml")

	if err != nil {
		return err
	}

	config.WriteString(yaml)

	cmd := exec.Command("kubectl", method, "-f", config.Name())
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = filepath.Dir(config.Name())

	// Set KUBECONFIG location
	cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%v", t.Kubeconfig))

	// Run the command
	err = cmd.Run()

	if err != nil {
		return err
	}

	if e := stderr.String(); e != "" {
		return errors.New(e)
	}

	return nil
}

func (t *TestCluster) Stop() error {
	if t.provider != nil {
		log.Println("🏁 Destroying cluster")

		return t.provider.Delete(t.Name, t.Kubeconfig)
	}

	return nil
}

var CurrentCluster TestCluster

func TestMain(m *testing.M) {
	CurrentCluster = TestCluster{}

	err := CurrentCluster.Start()

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// log.Println("🎁 Creating resources in cluster for testing")
	// err = CurrentCluster.ApplyBaselineConfig()

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Println("✅ Running tests")
	code := m.Run()

	err = CurrentCluster.Stop()

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	os.Exit(code)
}

const ClusterBaseline string = `
apiVersion: v1
kind: PersistentVolume
metadata:
  name: d1
  labels:
    type: local
spec:
  storageClassName: manual
  capacity:
    storage: 20Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/mnt/d1"
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: d2
  labels:
    type: local
spec:
  storageClassName: manual
  capacity:
    storage: 20Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/mnt/d2"
---
apiVersion: v1
kind: Service
metadata:
  name: wordpress-mysql
  labels:
    app: wordpress
spec:
  ports:
    - port: 3306
  selector:
    app: wordpress
    tier: mysql
  clusterIP: None
---
apiVersion: v1
kind: LimitRange
metadata:
  name: test-lr2
spec:
  limits:
  - max:
      cpu: "200m"
    min:
      cpu: "50m"
    type: Container
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-pv-claim
  labels:
    app: wordpress
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: apps/v1 # for versions before 1.9.0 use apps/v1beta2
kind: Deployment
metadata:
  name: wordpress-mysql
  labels:
    app: wordpress
spec:
  selector:
    matchLabels:
      app: wordpress
      tier: mysql
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: wordpress
        tier: mysql
    spec:
      containers:
      - image: mysql:5.6
        name: mysql
        env:
        - name: MYSQL_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mysql-pass
              key: password
        ports:
        - containerPort: 3306
          name: mysql
        volumeMounts:
        - name: mysql-persistent-storage
          mountPath: /var/lib/mysql
      volumes:
      - name: mysql-persistent-storage
        persistentVolumeClaim:
          claimName: mysql-pv-claim
---
apiVersion: autoscaling/v1
kind: HorizontalPodAutoscaler
metadata:
  name: wordpress-mysql-as
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: wordpress-mysql
  minReplicas: 1
  maxReplicas: 3
  targetCPUUtilizationPercentage: 50
---
apiVersion: v1
kind: Service
metadata:
  name: wordpress
  labels:
    app: wordpress
spec:
  ports:
    - port: 8088
  selector:
    app: wordpress
    tier: frontend
  type: LoadBalancer
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: wordpress-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - http:
      paths:
      - path: /foo
        pathType: Prefix
        backend:
          service:
            name: wordpress
            port:
              number: 8088
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: wp-pv-claim
  labels:
    app: wordpress
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: apps/v1 # for versions before 1.9.0 use apps/v1beta2
kind: Deployment
metadata:
  name: wordpress
  labels:
    app: wordpress
spec:
  selector:
    matchLabels:
      app: wordpress
      tier: frontend
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: wordpress
        tier: frontend
    spec:
      containers:
      - image: wordpress:4.8-apache
        name: wordpress
        env:
        - name: WORDPRESS_DB_HOST
          value: wordpress-mysql
        - name: WORDPRESS_DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mysql-pass
              key: password
        resources:
          limits:
            cpu: 200m
          requests:
            cpu: 200m
        ports:
        - containerPort: 80
          name: wordpress
        volumeMounts:
        - name: wordpress-persistent-storage
          mountPath: /var/www/html
      volumes:
      - name: wordpress-persistent-storage
        persistentVolumeClaim:
          claimName: wp-pv-claim
---
# Example replication controller. These are old school and basically replaced
# by deployments
apiVersion: v1
kind: ReplicationController
metadata:
  name: nginx
spec:
  replicas: 1
  selector:
    app: nginx
  template:
    metadata:
      name: nginx
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: pods-high
spec:
  hard:
    cpu: "1000"
    memory: 200Gi
    pods: "10"
  scopeSelector:
    matchExpressions:
    - operator : In
      scopeName: PriorityClass
      values: ["high"]
---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: pods-medium
spec:
  hard:
    cpu: "10"
    memory: 20Gi
    pods: "10"
  scopeSelector:
    matchExpressions:
    - operator : In
      scopeName: PriorityClass
      values: ["medium"]
---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: pods-low
spec:
  hard:
    cpu: "5"
    memory: 10Gi
    pods: "10"
  scopeSelector:
    matchExpressions:
    - operator : In
      scopeName: PriorityClass
      values: ["low"]
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd-elasticsearch
  labels:
    k8s-app: fluentd-logging
spec:
  selector:
    matchLabels:
      name: fluentd-elasticsearch
  template:
    metadata:
      labels:
        name: fluentd-elasticsearch
    spec:
      containers:
      - name: fluentd-elasticsearch
        image: quay.io/fluentd_elasticsearch/fluentd:v2.5.2
        resources:
          limits:
            memory: 200Mi
          requests:
            cpu: 50m
            memory: 200Mi
        volumeMounts:
        - name: varlog
          mountPath: /var/log
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
      terminationGracePeriodSeconds: 30
      volumes:
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
---
# Example stateful set
apiVersion: v1
kind: Service
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  ports:
  - port: 8089
    name: web
  clusterIP: None
  selector:
    app: nginx
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  serviceName: "nginx"
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: k8s.gcr.io/nginx-slim:0.8
        ports:
        - containerPort: 80
          name: web
        volumeMounts:
        - name: www
          mountPath: /usr/share/nginx/html
        resources:
          limits:
            cpu: 50m
          requests:
            cpu: 50m
  volumeClaimTemplates:
  - metadata:
      name: www
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 1Gi
---
# job example
apiVersion: batch/v1
kind: Job
metadata:
  name: pi
spec:
  template:
    spec:
      containers:
      - name: pi
        image: perl
        command: ["perl",  "-Mbignum=bpi", "-wle", "print bpi(2000)"]
      restartPolicy: Never
  backoffLimit: 4
  parallelism: 3
---
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: hello
spec:
  schedule: "*/1 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: hello
            image: busybox
            args:
            - /bin/sh
            - -c
            - date; echo Hello from the Kubernetes cluster
          restartPolicy: OnFailure

`
