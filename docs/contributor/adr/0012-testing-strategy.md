# Testing strategy of the APIGateway module

## Status
Accepted

## Context
The APIGateway module requires a testing strategy to ensure that the module is always functioning
as expected on all supported platforms.
However, running tests on all supported platforms in all cases is both time-consuming and expensive.
Therefore, we need to decide on a testing strategy that balances the need for comprehensive
testing with the requirement for fast feedback and development.

## Decision

The testing strategy for the APIGateway module will be implemented according to the following guidelines:
1. Tests that depend on the Gardener platform will not run on Pull Requests (PRs).
2. Gardener related tests will run during post-merge workflows and on scheduled runs.
3. If a Gardener-related test fails, the cluster will remain alive for debugging purposes.
4. In the event of a test failure during a post-merge workflow, the PR owner is responsible for resolving the issue.
5. PR tests should generally avoid relying on external resources.
This especially means that secrets should not be required for running PR tests whenever possible.
6. Integration tests for PRs will run on a local Kubernetes cluster using the k3d platform.
7. Compatibility and UI tests will run only on scheduled runs.
8. Tests ensuring release stability and readiness will be triggered during the release workflow.

Additionally, the following naming conventions will be adopted for workflows:
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

Tests will also be renamed to align with the adopted naming conventions.
