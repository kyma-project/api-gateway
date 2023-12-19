# Istio Performance Tests

This directory contains the scripts for running kyma performance tests.

## Test Setup

- Deploy a Kubernetes cluster with Kyma on a production profile
- Export the following variable:

```sh
export DOMAIN=<YOUR_CLUSTER_DOMAIN>
```

- Deploy helm chart to start load-testing

```sh
helm dependency update operator/performance_tests/load-testing/.
helm install goat-test --set domain="$DOMAIN" --create-namespace -n load-test operator/performance_tests/load-testing/.
```

- Run from main directory:

```sh
make perf-test
```
