# Tests

## Overview

This project contains API Gateway integration tests for Kyma Dashboard.

## Prerequisites

If you want to use an existing cluster, you must copy your cluster's kubeconfig file to `fixtures/kubeconfig.yaml`.

## Installation

To install the dependencies, run the `npm clean install` command.

## Test Development Using Headless Mode with Chrome Browser

### Using a Local Kyma Dashboard Instance

Start a k3d cluster:

```bash
npm run start-k3d
```

Start the local Kyma dashboard instance:

```bash
npm run start-dashboard
```

#### Run the Tests

```bash
npm run test
```

#### Run Cypress UI Tests in the Test Runner Mode

```bash
npm run start
```

### Using a Remote Kyma Dashboard Instance

#### Run Tests
```bash
CYPRESS_DOMAIN={REMOTE_CLUSTER_DASHBOARD_DOMAIN} npm run test
```

#### Run Cypress UI Tests in the Test Runner Mode

```bash
CYPRESS_DOMAIN={REMOTE_CLUSTER_DASHBOARD_DOMAIN} npm run start
```

## Run Tests in Continuous Integration System

Start a k3d cluster and run the tests:

```bash
./scripts/k3d-ci-kyma-dashboard-integration.sh
```
