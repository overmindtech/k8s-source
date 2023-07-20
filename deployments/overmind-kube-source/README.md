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

## Releasing

These charts are released automatically using [helm-chart-releaser](https://github.com/marketplace/actions/helm-chart-releaser). Chart version should match tags in the repo
