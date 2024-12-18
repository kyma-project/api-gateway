# Testing strategy of the APIGateway module

## Background
The APIGateway module requires a testing strategy to ensure that the module is always functioning
as expected on all supported platforms.
However, running tests on all supported platforms in all cases is both time-consuming and expensive.
Therefore, we need to decide on a testing strategy that balances the need for comprehensive
testing with the requirement for fast feedback and development.

## Testing strategy outline

The testing strategy for the APIGateway module will be implemented according to the following guidelines:
1. Tests that depend on the Gardener platform will not run on Pull Requests (PRs).
2. Gardener related tests will run during post-merge workflows and on scheduled runs.
3. If a Gardener-related test fails, the cluster will remain alive for debugging purposes.
4. In the event of a test failure during a post-merge workflow, the PR owner is responsible for resolving the issue.
5. PR tests should generally avoid relying on external resources.
This especially means that secrets should not be required for running PR tests whenever possible.
This allows for more secure testing and reduces the risk of leaking secrets.
6. Integration tests for PRs will run on a local Kubernetes cluster using the k3d platform.
7. Compatibility and UI tests will run only on scheduled runs.
8. Tests ensuring release stability and readiness will be triggered during the release workflow.
9. Generally, we do not want to test the functionality of the dependencies,
but rather make sure that the configuration for the used resources is correct.

Additionally, the following naming conventions will be adopted for workflows:
- Workflows that run before merge should be prefixed with `pull`.
- Workflows triggered after merge should be prefixed with `post`.
- Workflows running on schedule should be prefixed with `schedule`.
- Workflows related to release should be prefixed with `release`.
- Workflows that run on manual trigger will be prefixed with `call`.

## Rationale

The primary objective is to ensure that the module remains stable and prepared for release,
while providing prompt feedback for PRs.
This testing strategy prioritizes comprehensive e2e evaluations during post-merge and release pipeline stages,
while PR-level assessments focus on unit tests and fundamental integration tests.

Testing on Gardener clusters incurs significant computational and time costs.
End-to-end (e2e) Kubernetes scenarios are instead evaluated using k3d where feasible,
due to its reduced initialization time and higher reliability in controlled environments.
The probability of Gardener-specific test failures when k3d-based tests pass is considered low.
In the event of Gardener-related test failures after a PR has been merged,
the responsibility for stabilizing the associated workflows lies with the author of the merged PR.

As generally the tests do not rely on specific cloud provider configurations,
we perform Gardener-related tests mainly on AWS cloud,
as this is the most common cloud provider for BTP Kyma Runtime.

To ensure stability of image building, PRs build their own local image,
with the post-merge workflow using the image-builder image.
As the last point to catch any issues before the release,
the release workflow will run all tests, including compatibility, performance, and UI tests.

## Separation of concerns

To achieve the desired testing strategy,
the tests will be separated into three categories,
that are defined as follows:
- `unit tests` that run without any external dependencies, and do not require any pre-existing resources.
- `integration tests` that require external dependencies,
  but do not require a full Kubernetes cluster (for example, they might run on controller-runtime `envtest` environment).
- `lighweight e2e tests` that require a Kubernetes cluster but can run on a local cluster (k3d).
- `full e2e tests` that require a full production-like Kubernetes cluster (Gardener).

Additionally, the tests will be separated into the following groups:
- `compatibility tests` that ensure compatibility with Kubernetes versions.
- `performance tests` that ensure the performance of the module.
- `UI tests` that ensure the UI of the module, presented in the Kyma Dashboard is working as expected.
- `upgrade tests` that ensure the module can be upgraded without any issues.
- `zero downtime tests` that ensure the module can be upgraded without any downtime.

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
