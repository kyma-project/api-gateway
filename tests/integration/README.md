# Api-gateway component tests

Api-gateway component tests use the [cucumber/godog](https://github.com/cucumber/godog) library.

## Prerequisites

- Kubernetes cluster provisioned and KUBECONFIG set to point to it
- Kyma installed

### Environment variables

These environment variables determine how the tests are run on both Prow and your local machine:

- `EXPORT_RESULT` - set this environment variable to `true` if you want to export test results to JUnit XML, Cucumber JSON, and HTML report. The default value is `false`.

Customize `env_vars.sh` if necessary.

## Usage for standard API Gateway test suite

To start the test suite, run:

```
make test-integration
```
