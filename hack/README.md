# `hack/` Directory

This directory contains scripts and supporting files used during development and CI/CD.

Scripts directly under `hack/` are general-purpose developer utilities that can also be useful outside of CI. Scripts under `hack/ci/` are intended to be run exclusively by the CI environment and typically require CI-specific credentials and environment variables.

## Named Configurations

Both `hack/ci/k3d/` and `hack/ci/gardener/` use a **named configuration** convention to select a test environment preset. Each preset lives in a subdirectory:

```
hack/ci/k3d/configurations/<name>/vars.sh
hack/ci/gardener/configurations/<name>/vars.sh
hack/ci/gardener/configurations/<name>/shoot.yaml
```

The configuration is selected by setting `K3D_CONFIGURATION` or `GARDENER_CONFIGURATION` to the preset name (e.g. `default`, `gcp-ipv4`, `aws-dualstack`). The `vars.sh` file defines all environment-specific variables for that preset and they are auto-exported into the shell. For Gardener configurations, `shoot.yaml` is a cluster template whose `$VAR` references are filled in from those variables via `envsubst`.

This keeps environment differences (cloud provider, IP stack, node settings, etc.) entirely inside the configuration directory, while the scripts themselves stay generic.

## `common.sh`

All `hack/ci/` scripts source `hack/ci/common.sh` at startup. It provides shared utilities used across all CI scripts — argument and environment variable validation, configuration loading, and structured log output.
