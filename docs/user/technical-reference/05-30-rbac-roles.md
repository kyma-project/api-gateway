# API Gateway RBAC Roles

The API Gateway module provides predefined Kubernetes ClusterRoles to control access to its custom resources. These roles follow
the standard Kubernetes RBAC aggregation pattern and are automatically aggregated into the built-in `admin`, `edit`, and `view`
cluster roles.

## Overview

| Role Name                | Aggregates To | Description                                                                                                  |
|--------------------------|---------------|--------------------------------------------------------------------------------------------------------------|
| `kyma-api-gateway-admin` | `admin`       | Full access to all API Gateway resources                                                                     |
| `kyma-api-gateway-edit`  | `edit`        | Full access to `gateway.kyma-project.io` resources; read-only access to `operator.kyma-project.io` resources |
| `kyma-api-gateway-view`  | `view`        | Read-only access to all API Gateway resources                                                                |

## Roles

### Admin Role (`kyma-api-gateway-admin`)

The admin role grants full management permissions to all API Gateway custom resources. It aggregates into the Kubernetes
built-in `admin` ClusterRole.

**Permissions for `gateway.kyma-project.io` resources:**

| Resource            | Verbs                                                         |
|---------------------|---------------------------------------------------------------|
| `apirules`          | `create`, `delete`, `get`, `list`, `patch`, `update`, `watch` |
| `ratelimits`        | `create`, `delete`, `get`, `list`, `patch`, `update`, `watch` |
| `apirules/status`   | `get`                                                         |
| `ratelimits/status` | `get`                                                         |

**Permissions for `operator.kyma-project.io` resources:**

| Resource             | Verbs                                                         |
|----------------------|---------------------------------------------------------------|
| `apigateways`        | `create`, `delete`, `get`, `list`, `patch`, `update`, `watch` |
| `apigateways/status` | `get`                                                         |

---

### Editor Role (`kyma-api-gateway-edit`)

The editor role grants full management permissions to `gateway.kyma-project.io` resources, and read-only access to
`operator.kyma-project.io` resources. It aggregates into the Kubernetes built-in `edit` ClusterRole.

**Permissions for `gateway.kyma-project.io` resources:**

| Resource            | Verbs                                                         |
|---------------------|---------------------------------------------------------------|
| `apirules`          | `create`, `delete`, `get`, `list`, `patch`, `update`, `watch` |
| `ratelimits`        | `create`, `delete`, `get`, `list`, `patch`, `update`, `watch` |
| `apirules/status`   | `get`                                                         |
| `ratelimits/status` | `get`                                                         |

**Permissions for `operator.kyma-project.io` resources:**

| Resource             | Verbs                  |
|----------------------|------------------------|
| `apigateways`        | `get`, `list`, `watch` |
| `apigateways/status` | `get`                  |

---

### Viewer Role (`kyma-api-gateway-view`)

The viewer role grants read-only access to all API Gateway custom resources. It aggregates into the Kubernetes built-in
`view` ClusterRole.

**Permissions for `gateway.kyma-project.io` resources:**

| Resource            | Verbs                  |
|---------------------|------------------------|
| `apirules`          | `get`, `list`, `watch` |
| `ratelimits`        | `get`, `list`, `watch` |
| `apirules/status`   | `get`                  |
| `ratelimits/status` | `get`                  |

**Permissions for `operator.kyma-project.io` resources:**

| Resource             | Verbs                  |
|----------------------|------------------------|
| `apigateways`        | `get`, `list`, `watch` |
| `apigateways/status` | `get`                  |
