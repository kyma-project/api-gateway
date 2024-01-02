# Scripts for API Gateway Integration Tests

Scripts are based on the gardener provision scripts and utilities from `test-infra`:
https://github.com/kyma-project/test-infra/tree/main/prow/scripts

## Structure

- `gardener`: scripts for generic gardener de-/provision clusters and IaaS specifics in additional script files
- `jobguard`: job guard script that executes `/prowutils/jobguard` utility from `test-infra/kyma-integration` image, waiting for image build job to complete
- `lib`: general utility functions for logging, kyma cli install, etc.

## Entrypoint for Integration Tests on Gardener

API Gateway integration tests are triggered on `presubmit` (PR) and `postsubmit` (main branch) by executing `integration-gardener.sh` script, which provisions a Gardener cluster and starts the tests in `integration-tests.sh`.

All required environment variables are provided by Prow.
