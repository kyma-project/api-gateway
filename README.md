[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/api-gateway)](https://api.reuse.software/info/github.com/kyma-project/api-gateway)

# API Gateway

API Gateway is a Kyma module with which you can expose and secure APIs.

The API Gateway module offers the following features:

- API Exposure: The module uses Istio features to help you easily and securely expose your workloads by creating APIRule custom resources. With an APIRule, you can perform the following actions:
  - Group multiple workloads and expose them under a single host.
  - Use a short host name to simplify the migration of resources to a new cluster.
  - Configure the **noAuth** access strategy, which offers a simple configuration to allow access to specific HTTP methods.
  - Secure your workloads by configuring **jwt** or **extAuth** access strategies. The **jwt** access strategy enables you to use Istio's JWT configuration to protect your exposed services and interact with them using JSON Web Tokens. The **extAuth** access strategy allows you to implement custom authentication and authorization logic.

- Gateway configuration:
  - Default Kyma Gateway: The module sets up the default TLS Kyma Gateway, which uses the default domain and a self-signed certificate.
  - Custom Gateway: The module allows you to configure a custom Gateway, which is recommended for production environments. Additionally, it enables you to expose workloads using a custom domain and DNSEntry. 

- Rate Limiting: The module simplifies local rate limiting on the Istio service mesh layer. You can configure it using a straightforward RateLimit custom resource.

## Installation

### Prerequisites

- [k3d](https://k3d.io/stable/)

### Procedure

1. Create a Kyma cluster.

      ```bash
      k3d cluster create kyma --port 80:80@loadbalancer --port 443:443@loadbalancer --k3s-arg "--disable=traefik@server:*"
      ```

2. Create the `kyma-system` namespace with enabled Istio sidecar injection.

      ```bash
      kubectl create ns kyma-system
      kubectl label namespace kyma-system istio-injection=enabled --overwrite
      ```

2. To use the API Gateway module, you must also add the Istio module. Run:

      ```bash
      kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
      kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml
      ```

      To verify if the Istio module is added, check the state of the Istio CR:

      ```bash
      kubectl get istios/default -n kyma-system
      ```

      If successful, you get the following output:

      ```bash
      NAME      STATE
      default   Ready
      ```

3. Add the API Gateway module.

      ```bash
      kubectl apply -f https://github.com/kyma-project/api-gateway/releases/latest/download/api-gateway-manager.yaml
      kubectl apply -f https://github.com/kyma-project/api-gateway/releases/latest/download/apigateway-default-cr.yaml
      ```

      To verify if the API Gateway module is added, check the state of the API Gateway CR:

      ```bash
      kubectl get apigateways/default -n kyma-system
      ```

      If successful, you get the following output:

      ```bash
      NAME      STATE
      default   Ready
      ```

For more installation options, see the [installation guide](./docs/contributor/01-00-installation.md).

## Repository Conventions

This repository uses two metadata files that drive CI/CD automation. They should not be changed by developers working on forks unless explicitly noted.

### `MINOR_VERSION`

Contains the current major.minor version of the module (e.g. `3.11`). It represents what this branch *is*, not what it will become:

- On `main`: the version currently in development (e.g. `3.11` means the next release will be `3.11.x`)
- On a `release-X.Y` branch: always `X.Y`, set when the branch was created and never changed
- On feature/bugfix branches: inherited from the branch they were cut from — no changes needed

When a new minor release is prepared, the `prepare-new-minor` workflow branches off `release-X.Y`, then bumps `MINOR_VERSION` on `main` to the next minor via a PR.

### `RELEASE_REPOSITORY`

Contains the canonical GitHub repository where official module releases are published (e.g. `kyma-project/api-gateway`). It is used by:

- `hack/ci/get-reference-release.sh` — to find the official release to upgrade from in upgrade tests
- `make deploy-release` — to install a specific release version into a cluster

**Forks should not change this file.** A developer working on a fork still upgrades from and installs the official upstream releases. Only the repository that owns the official release process should have its own name here.

### `hack/` Directory

The `hack/` directory contains scripts and supporting files used during development and CI/CD. Scripts directly under `hack/` are general-purpose developer utilities that can also be useful outside of CI. Scripts under `hack/ci/` are intended to be run exclusively by the CI environment and typically require CI-specific credentials and environment variables.

#### Named Configurations

Both `hack/ci/k3d/` and `hack/ci/gardener/` use a **named configuration** convention to select a test environment preset. Each preset lives in a subdirectory:

```
hack/ci/k3d/configurations/<name>/vars.sh
hack/ci/gardener/configurations/<name>/vars.sh
hack/ci/gardener/configurations/<name>/shoot.yaml
```

The configuration is selected by setting `K3D_CONFIGURATION` or `GARDENER_CONFIGURATION` to the preset name (e.g. `default`, `gcp-ipv4`, `aws-dualstack`). The `vars.sh` file defines all environment-specific variables for that preset and they are auto-exported into the shell. For Gardener configurations, `shoot.yaml` is a cluster template whose `$VAR` references are filled in from those variables via `envsubst`.

This keeps environment differences (cloud provider, IP stack, node settings, etc.) entirely inside the configuration directory, while the scripts themselves stay generic.

#### `common.sh`

All `hack/ci/` scripts source `hack/ci/common.sh` at startup. It provides shared utilities used across all CI scripts — argument and environment variable validation, configuration loading, and structured log output.

## Useful Links

To learn how to use the API Gateway module, read the documentation in the [`user`](./docs/user/) directory.

If you are interested in the detailed documentation of the Kyma API Gateway Operator's design and technical aspects, check the [`contributor`](./docs/contributor/) directory.

## Contributing
<!--- mandatory section - do not change this! --->

See the [Contributing](CONTRIBUTING.md) guidelines.

## Code of Conduct
<!--- mandatory section - do not change this! --->

See the [Code of Conduct](CODE_OF_CONDUCT.md) document.

## Licensing
<!--- mandatory section - do not change this! --->

See the [license](./LICENSE) file.
