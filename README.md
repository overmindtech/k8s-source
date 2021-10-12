# Kubernetes Source

This primary source is designed to be used with [srcman](https://github.com/dylanratcliffe/srcman). It connects directly to your Kubernetes clusters using the k8s API and then responds to [SDP](https://github.com/dylanratcliffe/sdp) requests over a NATS network. Usually this will be run as a container and managed by [srcman](https://github.com/dylanratcliffe/srcman).

## Config



## Search

The backends in this package implement the `Search()` method. The query format that they are expecting is a JSON object with one or more of the following keys, with strings in the corresponding string format:

* `labelSelector`: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
* `fieldSelector`: https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/

An example would be:

```json
{
    "labelSelector": "app=wordpress"
}
```

or

```json
{
    "labelSelector": "environment=production,tier!=frontend",
    "fieldSelector": "metadata.namespace!=default"
}
```

Other fields can also be set of advanced querying is required, these fields must match the JSON schema for `ListOptions`: https://pkg.go.dev/k8s.io/apimachinery@v0.19.2/pkg/apis/meta/v1#ListOptions

## Development

### Testing

The tests for this package rely on having a Kubernetes cluster to interact with. This is handled using [kind](https://github.com/kubernetes-sigs/kind) when the tests are started. Please make sure that you have the required software installed:

* [kind](https://github.com/kubernetes-sigs/kind)
* [kubectl](https://kubernetes.io/docs/tasks/tools/)
* [Docker](https://docs.docker.com/get-docker/)
