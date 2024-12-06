# Testing strategy of the APIGateway module

## Status
Accepted

## Context
The APIGateway module requires a testing strategy to ensure that the module is functioning as expected on all supported platforms.
However, running tests on all supported platforms is time-consuming and expensive.
Therefore, we need to decide on a testing strategy that balances the need for comprehensive testing with the need for fast feedback.

## Decision

The testing strategy for the APIGateway module will be implemented according to the following guidelines:
1. Tests that depend on the Gardener platform will not run on Pull Requests (PRs).
2. Gardener related tests will run on post-merge workflows and scheduled runs.
3. Failures of Gardener related tests will result in the cluster being kept alive for debugging purposes.
4. In case a test fails on post-merge workflow, the PR owner is responsible for fixing the issue.
5. PRs should generally not require external resources to run tests.
This especially means that secrets should not be required to run tests on PRs if possible.
6. Integration tests that are performed on PRs will run on a Kubernetes cluster local to the runner, using k3d platform.
7. Compatibility and UI tests will run only on scheduled runs.
8. Tests that ensure release stability and readiness will run when a release workflow is triggered.

In addition, naming conventions for the workflows will be adopted:
- Workflows that run before merge should be prefixed with `pull`.
- Workflows triggered after merge should be prefixed with `post`.
- Workflows running on schedule should be prefixed with `schedule`.
- Workflows related to release should be prefixed with `release`.
- Workflows that run on manual trigger will be prefixed with `call`.

## Consequences

The module will adopt the test run strategy according to the following matrix:

| Trigger/Job                                                        | lint | unit tests | integration tests | custom domain int test | upgrade tests | compatibility test | UI tests | APIRule Migration Zero downtime test |
|--------------------------------------------------------------------|------|------------|-------------------|------------------------|---------------|--------------------|----------|--------------------------------------|
| PR (own image, all on k3d)                                         | x    | x          | x (k3d)           |                        |               |                    |          |                                      |
| main (image-builder image)                                         | x    | x          | x (k3d, AWS)      | x (AWS, GCP)           | x (k3d)       |                    |          | x (k3d, AWS)                         |
| PR to release branches (own image)                                 | x    | x          | x (k3d)           |                        |               |                    |          |                                      |
| schedule (image-builder image)                                     |      |            | x (k3d, AWS)      | x (AWS, GCP)           | x (k3d)       | x (k3d, AWS)       | x (k3d)  | x (k3d, AWS)                         |
| release (create release workflow) (image-builder image - prod art) | x    | x          | x (k3d, AWS)      | x (AWS, GCP)           | x (k3d)       |                    |          | x (k3d, AWS)                         |

Additionally, the tests will be re-named according to the naming conventions decided upon.
