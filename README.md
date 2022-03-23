# Kubernetes Source

This primary source is designed to be used with [srcman](https://github.com/overmindtech/srcman). It connects directly to your Kubernetes clusters using the k8s API and then responds to [SDP](https://github.com/overmindtech/sdp) requests over a NATS network. Usually this will be run as a container and managed by [srcman](https://github.com/overmindtech/srcman).

## Config

All configuration options can be provided via the command line or as environment variables:

| Environment Variable | CLI Flag | Description |
|----------------------|----------|-------------|
| `CONFIG`| `--config` | config file (default is $HOME/.k8s-source.yaml). Can be used instead of the CLI or environment variables if needed |
| `LOG`| `--log` | Set the log level. Valid values: panic, fatal, error, warn, info, debug, trace |
| `NATS_SERVERS`| `--nats-servers` | A list of NATS servers to connect to |
| `NATS_NAME_PREFIX`| `--nats-name-prefix` | A name label prefix. Sources should append a dot and their hostname .{hostname} to this, then set this is the NATS connection name which will be sent to the server on CONNECT to identify the client |
| `NATS_JWT` | `--nats-jwt` | ✅ | The JWT token that should be used to authenticate to NATS, provided in raw format e.g. `eyJ0eXAiOiJKV1Q{...}` |
| `NATS_NKEY_SEED` | `--nats-nkey-seed` | ✅ | The NKey seed which corresponds to the NATS JWT e.g. `SUAFK6QUC{...}` |
| `KUBECONFIG`| `--kubeconfig` | Path to the kubeconfig file containing cluster details (default: `/etc/srcman/config/kubeconfig`) |
| `MAX-PARALLEL`| `--max-parallel` | Max number of requests to run in parallel |

### `srcman` config

When running in srcman, most of the config is provided automatically. All you
need to provide is:

* `kubeconfig`: The contents of your kubeconfig file. This will be mounted at
  `/etc/srcman/config/kubeconfig` and loaded automatically
*  `k8s-source.yaml`: The contents of the YAML file containing config, the only
   thing you need to set in here is `max-parallel`

The following is an example of the configuration required for `srcman`:

```yaml
apiVersion: srcman.example.com/v0
kind: Source
metadata:
  name: source-sample
spec:
  image: ghcr.io/overmindtech/k8s-source:main
  replicas: 2
  manager: manager-sample
  config:
    k8s-source.yaml: |
      max-parallel: 24
    kubeconfig: |
      apiVersion: v1
      clusters:
      - cluster:
          certificate-authority: /etc/srcman/config/ca.crt
          extensions:
          - extension:
              last-update: Wed, 13 Oct 2021 17:02:23 BST
              provider: minikube.sigs.k8s.io
              version: v1.23.2
            name: cluster_info
          server: https://127.0.0.1:65526
        name: minikube
      contexts:
      - context:
          cluster: minikube
          extensions:
          - extension:
              last-update: Wed, 13 Oct 2021 17:02:23 BST
              provider: minikube.sigs.k8s.io
              version: v1.23.2
            name: context_info
          namespace: default
          user: minikube
        name: minikube
      current-context: minikube
      kind: Config
      preferences: {}
      users:
      - name: minikube
        user:
          client-certificate: /etc/srcman/config/client.crt
          client-key: /etc/srcman/config/client.key
    client.crt: |
      -----BEGIN CERTIFICATE-----
      MIIDITCCAgmgAwIBAgIBAjANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwptaW5p
      a3ViZUNBMB4XDTIxMTAxMjEzNTYyNloXDTIyMTAxMzEzNTYyNlowMTEXMBUGA1UE
      ChMOc3lzdGVtOm1hc3RlcnMxFjAUBgNVBAMTDW1pbmlrdWJlLXVzZXIwggEiMA0G
      CSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC/bs9yX6zRrDACuOnIKKmPH0HGNcJF
      X6dRxou3qSI2zCTTGM/20UmtBoWOpg5EbB8iVhnFD9hJ+XBstG5DDVlA7PqQ4fCK
      hjNsbethaNgPSkAj7qbiSvgbqouEUaamEryzlqZMwjGtLWC7s7PeXCOTvt8062Mr
      2VBtfSqemJQnvAdgZpiB131f8QnYTzGU6vySb7C6Z3T2IqhWIAWlWCkgyHtUmGtU
      0ikQ3B58FT63dmggLKL3djCExUsDm1dah36MOQjoP2cagWFJ6J154Of0r+grKyHe
      Bra13BTF7pYeJgx85ZO/N+oCpp9UU0D1QEe1KHVqVNY9jhD03W7W0EJRAgMBAAGj
      YDBeMA4GA1UdDwEB/wQEAwIFoDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUH
      AwIwDAYDVR0TAQH/BAIwADAfBgNVHSMEGDAWgBSpCTbPiww8DoqAO8BageGGkYgK
      tzANBgkqhkiG9w0BAQsFAAOCAQEAaeTz+gdzAqUFd4I/UVBYcjMkNoY4EPKPCrwD
      uRFQvJou2pzVBsZBVRzBIPiv5G2TRESI91eVeDzmgh1MYeM+VwMkPxWvyUUvR1yN
      1bhsMEUP/f+LXo9+o8GxYWPI1vgLL1ERNgR1waFloH9wElu41fuG6YyXzn5q8Crh
      CKZ3dviREbNZTdeEQYG4TZvDGK7FIUDJzFe8lw7vBmeLuQtiGsdSIjQVPuibqjgO
      zag1xcLx7laGr3wFUQRtIez5TdV2EznUUK4n/ZRCZvaKhTXJnhMyFcox6MkNV3nR
      kwlQQ4dmqbofuEasHRMtmuuPoMVb3REN+Y1e+/ojsN7DbUquqQ==
      -----END CERTIFICATE----- 
    client.key: |
      -----BEGIN RSA PRIVATE KEY-----
      MIIEowIBAAKCAQEAv27Pcl+s0awwArjpyCipjx9BxjXCRV+nUcaLt6kiNswk0xjP
      9tFJrQaFjqYORGwfIlYZxQ/YSflwbLRuQw1ZQOz6kOHwioYzbG3rYWjYD0pAI+6m
      4kr4G6qLhFGmphK8s5amTMIxrS1gu7Oz3lwjk77fNOtjK9lQbX0qnpiUJ7wHYGaY
      gdd9X/EJ2E8xlOr8km+wumd09iKoViAFpVgpIMh7VJhrVNIpENwefBU+t3ZoICyi
      93YwhMVLA5tXWod+jDkI6D9nGoFhSeideeDn9K/oKysh3ga2tdwUxe6WHiYMfOWT
      vzfqAqafVFNA9UBHtSh1alTWPY4Q9N1u1tBCUQIDAQABAoIBAAkO2DgEOOwu5pKq
      Zz12VxeTlgwn7QpVTVh8OY42LY1EOZXXfbejDYZnYZhvWQt5xjtcsZl2d3iAmgY6
      v2Di189Pp0eFuVkEophF1zZjvJ10mPZaS4E3pOfCORnIt0byagVhYnsNUUZteD9J
      cIBcAb7y8CLT5Hxlqv2TR5n7hD8g/HDA6u2zRs2tr0I0mrOQkx1N2LgjfmTDmy5K
      b/Yt9/FIswSisz7UNvYO1JeMA0CxkX3ssE4mVXmJyOjHH5LeAeStQIg43YXoWbf7
      yBoDQI+K6uE0k0hjP0QpsQIgzBWgY9qkLA8c82uT3YvkCcTKjxmTQHpal/M43wSL
      ch8wrGUCgYEAxlIKMlEzD5fdhwU/LbfsyGe93UNvxLdYYabjTG1FI/aiHokO2ED1
      ZnVED66hvCbGgnwETs2rJFtGUUbUgfuX9Y6FCMzmOPaoLgruNHbZg1qQiHOZGk2A
      mHeibTA30G4hP0GCqzLoawjQFMMp7dlRUbWwU/GKHcAtRp9QXWfW56sCgYEA9xvz
      6XuAeeuzArWvRPwCsCxEMT9O/JYh3tcPTEeEhZkMqcpShVR20sQQCactxLcW6LzT
      ht1gv7pGOVcRRKeJBda5ww9tgnJJAlixly0vTUGVzdV//hLqXgYCaW/BUlMbN3n1
      y/X6liFGeWp6Vpco1xyU8A32CVNKUNzmv9FQEfMCgYEAs9u7e57AnCeytL1Bawkf
      KTFMs9pxBwrwkL917N48kj0fEmpimCVxaZZ4P3C1JZpU9gnbLkzAJZzRzOxb1faC
      /iRe6nhJYufv5rHrDpGq+sGrytRrybr4IU5+dGACfnkileenxfPJbSj07Z+B6z/n
      zB7m53prNEgRx7a8f7mo4TkCgYBcjSmjz0/lWjQn1aiZq9HN7iZ0U4Pf8tMoxV/D
      cB3gc9xcU5zotyPx+OEQ3H616OU5sk9/ebbc2IWowEWFc0JM34mf101qyCc0K8gI
      GTJYOzJCb66KmMcTBCkvGF5N2TaeZp17ENwUEs50dz7u45q2Rsw5xODbyUhSVQpP
      2bOlpQKBgDH/NXG+5YYQCux4qdhCpwMbcG8NVA/5tM6L+Sa4aL9gCus5eX9QYt3N
      MPQo15kCRga4Ss+ZNty6Vf+noRRV5D73Cgi2/mTf6cvlFIQbmjxe+5V2Km7ibSY8
      Tn9U/FCBNTKoiCN/9cfjzKQBB6ieKyyy0xwKniwRVA0skf6raVC5
      -----END RSA PRIVATE KEY-----
    ca.crt: |
      -----BEGIN CERTIFICATE-----
      MIIDBjCCAe6gAwIBAgIBATANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwptaW5p
      a3ViZUNBMB4XDTIwMDkxODEzNDU1N1oXDTMwMDkxNzEzNDU1N1owFTETMBEGA1UE
      AxMKbWluaWt1YmVDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAL28
      bM8jWz2iR6XC6RI+SHJoZZ1yaA/iP9V1c8RsfeQDGMBtoHixEhK+VM7bG3LKi5k3
      wXua177/+p+6vu2JrOXBd2GetDv+7NxR6UoNziejnnvWGtEUx2RYYFNhcN6Wh7rd
      LNQu7h2lo8TPsLZPn7C5goMaJ5iUcA0nIQjZhOo7RQjGEl1BdsWorhjSSh0esIfu
      zUH4YDzO1s8WvjWQiPzyW77yAM5SwjrrmAPB0m2sS7jaSiEeANIlWCG8jlWYKv3+
      R2B16/3vUl8F03nmdsoVvfwyF1HoyVZZp1pCeYu4U/VPYygZOUa6WxCROEv+R7Yu
      4D8Es6W5cws62BAV4vsCAwEAAaNhMF8wDgYDVR0PAQH/BAQDAgKkMB0GA1UdJQQW
      MBQGCCsGAQUFBwMCBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQW
      BBSpCTbPiww8DoqAO8BageGGkYgKtzANBgkqhkiG9w0BAQsFAAOCAQEAR2w4HP1q
      x0M+4gAkkT4w8eD4qyU7SUVaKiXN8lAW3G9DODEynqQKXUbOGjvdFrX5gVbHeqn9
      9u2tSoxEg0s7yA80gwMYcIsHto3ICEGtys18YgJ/dJMbyAV5mcX2B5ge9mdjddBk
      sIVap+CXdZ5dv3OG9z8HYGFNcBsZX7Ef6UEQka6vX8qYoi4EXPe4jN3qPEJm34Jt
      PCXDGM+CwEEulJJ/MxUiY+UhgqQZyoI7LeTSwFF5K/gzBg6nJ9azCZHLSIYDixX/
      +irPLfBSalNM51ZxtY5oH7+m6cATMCAU9bOblDVOlK0HcDl2OP6soNU6+SU4u5kS
      G9lupxvmia+eug==
      -----END CERTIFICATE-----
```

### Health Check

The source hosts a health check on `:8080/healthz` which will return an error if NATS is not connected. An example Kubernetes readiness probe is:

```yaml
readinessProbe:
  httpGet:
    path: /healthz
    port: 8080
```

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

**IMPORTANT:** If you already have kubectl configured and are connected to a cluster, that cluster is what will be used for testing. Resources will be cleaned up with the exception of the testing namespace. If a cluster is not configured, or not available, one will be created (and destroyed) using `kind`. This behavior may change in the future as I see it being a bit risky as it could accidentally run the tests against a production cluster, though that would be a good way to validate real-world use-cases...
