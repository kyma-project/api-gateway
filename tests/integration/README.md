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

## Integration tests run on `presubmit` and `postsubmit` in Prow

Job definitions are specified [in test-infra repository](https://github.com/kyma-project/test-infra/blob/main/templates/data/api-gateway-validation.yaml).

## Usage for custom-domain test suite

### Set the custom domain environment variables

If you are using Gardener, make sure that your Kubernetes cluster has the `shoot-cert-service` and `shoot-dns-service` extensions enabled. The desired shoot specification is mentioned in the description of this [issue](https://github.com/kyma-project/control-plane/issues/875).
Obtain a service account access key with permissions to maintain custom domain DNS entries and export it as a JSON file. To learn how to do it, follow this [guide](https://cloud.google.com/iam/docs/keys-create-delete).

Set the following environment variables:
- `TEST_DOMAIN` - your default Kyma domain, for example, `c1f643b.stage.kyma.ondemand.com`
- `TEST_CUSTOM_DOMAIN` - your custom domain, for example, `custom.domain.build.kyma-project.io`
- `TEST_SA_ACCESS_KEY_PATH` - the path to the service account access key exported as a JSON file, for example, `/Users/user/gcp/service-account.json`

### Run the tests

To start the test suite, run:

```
make test-custom-domain
```
