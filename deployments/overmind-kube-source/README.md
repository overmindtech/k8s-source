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

These charts are released automatically using [helm-chart-releaser](https://github.com/marketplace/actions/helm-chart-releaser). Chart version should match tags in the repo. For some reason though the action only checks for changes since the last tag, so if you update the version, commit it and tag it, it sees no changes since the last tag and doesn't release it. This is nonsense. I'm assuming they want you to tag it, then update the version number in a subsequent commit. It's dumb but there you are.