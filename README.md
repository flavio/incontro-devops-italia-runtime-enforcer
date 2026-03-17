# incontro-devops-italia-demo

Live demo for Incontro DevOps Italia showcasing workload policy enforcement on a NATS deployment using Runtime Enforcer.

## Prerequisites

- [`kubectl`](https://kubernetes.io/docs/tasks/tools/)
- [`helm`](https://helm.sh/docs/intro/install/) with the [NATS chart repo](https://nats-io.github.io/k8s/helm/charts/) added:
  ```console
  helm repo add nats https://nats-io.github.io/k8s/helm/charts/
  ```
- [`bat`](https://github.com/sharkdp/bat) for syntax-highlighted YAML output
- A running Kubernetes cluster with Runtime Enforcer installed

## Run

```console
make demo
```
