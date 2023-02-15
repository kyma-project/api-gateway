# Api-gateway component tests

Api-gateway component tests use the [cucumber/godog](https://github.com/cucumber/godog) library.

## Prerequisites

- Kubernetes cluster provisioned and KUBECONFIG set to point to it
- Kyma installed
- OIDC issuer for JWT with claims
  - `scope` with values `read` and `write`
  - `aud` with values `https://example.com` and `https://example.com/user`

### Environment variables

These environment variables determine how the tests are run on both Prow and your local machine:

- `EXPORT_RESULT` - set this environment variable to `true` if you want to export test results to JUnit XML, Cucumber JSON, and HTML report. The default value is `false`.

Customize `env_vars.sh` if necessary.

To run these tests on your cluster you have to set those environment variables:

```
export KYMA_DOMAIN=<YOUR_KYMA_DOMAIN>
export CLIENT_ID="<YOUR_CLIENT_ID>"
export CLIENT_SECRET="<YOUR_CLIENT_SECRET>"
export OIDC_ISSUER_URL="<YOUR_OIDC_ISSUER_URL>"
```

## Usage for standard API Gateway test suite

To start the test suite, run:

```
make test-integration
```
