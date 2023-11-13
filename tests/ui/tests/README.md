# Tests

## Overview

This project contains API Gateway integration tests for Kyma Dashboard.

## Prerequisites

If you want to use an existing cluster, you must copy your cluster's kubeconfig file to `fixtures/kubeconfig.yaml`.

## Installation

To install the dependencies, run the `npm clean install` command.

## Test development using Headless mode with Chrome browser

### Using a local Kyma Dashboard instance

Start a k3d cluster:

```bash
npm run start-k3d
```

Start the local Kyma Dashboard instance:

```bash
npm run start-dashboard
```

#### Run tests

```bash
npm run test
```

#### Run Cypress UI tests in the test runner mode

```bash
npm run start
```

### Using a remote Kyma Dashboard instance

#### Optional: Log in to a cluster using OIDC

If a cluster requires OIDC authentication, include the additional arguments **CYPRESS_OIDC_PASS** and **CYPRESS_OIDC_USER** while running the npm scripts.

#### Run tests
```bash
CYPRESS_OIDC_PASS={YOUR_PASSWORD} CYPRESS_OIDC_USER={YOUR_USERNAME} CYPRESS_DOMAIN={REMOTE_CLUSTER_DASHBOARD_DOMAIN} npm run test
```

#### Run Cypress UI tests in the test runner mode

```bash
CYPRESS_OIDC_PASS={YOUR_PASSWORD} CYPRESS_OIDC_USER={YOUR_USERNAME} CYPRESS_DOMAIN={REMOTE_CLUSTER_DASHBOARD_DOMAIN} npm run start
```

## Run tests in Continuous Integration System

Start a k3d cluster and run the tests:

```bash
./scripts/k3d-ci-kyma-dashboard-integration.sh
```
