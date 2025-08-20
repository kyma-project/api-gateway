# End-to-End Tests

This directory contains definitions and implementations of end-to-end tests for API Gateway Operator.

## Running End-to-End Tests

E2E tests run against a cluster that is set up as active in your `KUBECONFIG` file. If you want to change the target,
export an environment variable `KUBECONFIG` to specific kubeconfig file. Make sure you have admin permission on the
cluster.

To run the E2E tests, use the `make e2e-test` command in the project's root directory.
You can also set the IMG variable to specify the API Gateway Operator image to test.

## Writing End-to-End Tests

### Helper Functions

When providing a helper function (part of `test/e2e/pkg/helpers`),
please ensure that it follows the following guidelines:

**Arguments and return values**
1. The function **MAY** have an associated `Options` type if it has optional arguments.
   The `Options` type **SHOULD** be configurable with functional options.
2. The function **MUST** follow argument and return structure as follows:
   - Arguments: `t *testing.T, [...], options ...Option`
   - Returns: `[...], error`
   Where `[...]` is a list of required arguments or return values.
3. Arguments that are optional **MUST** be put in the `Options` type.
4. The function **MAY** assume that options contains at most one element,
   but **MUST** check that it is not empty.
5. The function **MUST** use the `t.Context()` as context for any operations that require a context.
   In case there is a need to extend the context, (e.g. to add a timeout), it **MUST** build it on top of `t.Context()`.

**Error handling**
1. The function **MUST** return an error if it fails to complete its task.
2. There **MUST NOT** be any assertion present that would cause the test to fail.
   Example: `assert.NoError(t, err)` or `t.Fail()` is not allowed.

**Cleanup**

1. If the function creates any resources, it **MUST** declare a cleanup function
   that will clean up those resources after the test completes.
   This is done by calling `setup.DeclareCleanup` function from `test/e2e/pkg/setup`.
2. The cleanup function **MUST NOT** return an error.
3. In case an error occurs during cleanup, it **SHOULD** be logged using `t.Logf`.
4. The cleanup function **MUST** use the `setup.GetCleanupContext` from `test/e2e/pkg/setup` in case it requires a context.

**Logging**
1. The function **MUST** use `t.Log` or `t.Logf` for logging.
2. The function **SHOULD** log the start and end of the operation.
3. The function **SHOULD** log any errors that occur during the operation.
4. Performing of any major operation (e.g. creating a resource) **SHOULD** generally be logged.

### Test cases

When writing a test case, please ensure that it follows the following guidelines:

1. If multiple test cases are defined with the same background, they **MUST** be called using a `t.Run` subtest.
2. Setup of the test background (e.g., creating a cluster, installing Istio Operator) **MUST** be done outside the `t.Run` subtest.
3. Tests **MAY** use the assertions `assert` or `require` from the `github.com/stretchr/testify` package.
4. Related `yaml` files **SHOULD** be placed in the same directory as the test case, and **SHOULD** generally be loaded
   with `//go:embed` directive.
5. Tests **SHOULD** generally have as little logic and setup in them as possible.
   The test should be easily readable, and logic (e.g. generation of random names, creating test Pods) should be moved to helper functions.
