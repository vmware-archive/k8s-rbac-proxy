# Kubernetes RBAC Proxy

## Objective

to isolate controller effects to one or *more* namespaces

- pro: allows to decouple number of controllers from number of namespaces managed by these controllers
- pro: avoids duplication of namespace configuration in rbac and controller configs
- pro: works with controllers that cannot be modified or do not support namespacing out of the box
- con: ideally would be implemented in kubernetes api server
- con: proxying overhead?
- con: controller configuration has to be modified to redirect api requests (via env variable)

## Architecture

```
+-------------------------+             +-------+             +----------------------------+
| controller (downstream) | --- TLS --> | proxy | --- TLS --> | kube api-server (upstream) |
+-------------------------+             +-------+             +----------------------------+
```

- both controller and proxy would typically run inside the cluster
- TLS certs are issued through Kubernetes CA

## Alternative solutions

- modify controller source code to support multiple namespaces
  - [coreos/prometheus-operator](https://github.com/coreos/prometheus-operator) seems to achieve that with their own library: [pkg/listwatch](https://github.com/coreos/prometheus-operator/tree/f08c5ac5e9e74890db5035e7ea26a365edc2bff7/pkg/listwatch) -> [example usage](https://github.com/coreos/prometheus-operator/blob/f08c5ac5e9e74890db5035e7ea26a365edc2bff7/pkg/prometheus/operator.go#L209-L219)

## Docs

To install see [./hack/deploy.sh](./hack/deploy.sh).

- [Talk abstract](docs/talk-abstract.md)
- Development
  - [Example list response](docs/example-list-resp.md)
  - [Example watch response](docs/example-watch-resp.md)

## Use cases

- [etcd-operator: Support cross-namespace cluster operation](https://github.com/coreos/etcd-operator/issues/859)
  - [Add option to act as cluster wide](https://github.com/coreos/etcd-operator/pull/1777)

## TODO

- list: implement limit & continue token support
- list: implement list's revisionVersion support
- deletecollection

## Previously Seen Errors

Do let's know if you run into them.

```
build-controller-ff68c9946-ftgnr > build-controller | W0115 01:22:22.955946       1 reflector.go:341] github.com/knative/build/pkg/client/informers/externalversions/factory.go:114: watch of *v1alpha1.Build ended with: very short watch: github.com/knative/build/pkg/client/informers/externalversions/factory.go:114: Unexpected watch close - watch lasted less than a second and no items received

build-controller-ff68c9946-ftgnr > build-controller | W0115 01:22:50.191934       1 reflector.go:341] github.com/knative/build/vendor/github.com/knative/caching/pkg/client/informers/externalversions/factory.go:117: watch of *v1alpha1.Image ended with: very short watch: github.com/knative/build/vendor/github.com/knative/caching/pkg/client/informers/externalversions/factory.go:117: Unexpected watch close - watch lasted less than a second and no items received
```
