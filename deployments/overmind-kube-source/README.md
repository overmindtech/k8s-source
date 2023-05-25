# K8s Source Helm Chart

## Developing

Installing into a local cluster:

```
helm install k8s-source deployments/overmind-kube-source --set source.natsJWT=REPLACEME,source.natsNKeySeed=REPLACEME
```

Removing the chart:

```
helm uninstall k8s-source
```
