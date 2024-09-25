# Geoblocking

## Status

Draft

## Context

Geoblocking is a feature allowing to block incoming traffic coming from certain IP addresses, that are exclusive to certain countries or regions.
Can be utilized against anonymous users and system to system network communication, where both are identified only by its source IP address.
Many companies are using geoblocking to protect their services from unwanted traffic, such as DDoS attacks, or to comply with legal regulations.
Therefore, creating this kind of feature is a convenience for the user, as it extracts the need to implement it on its own. 

## Decision

Istio allows to delegate the access control to an external authorization service, which may be just a regular Kubernetes service. This seems to be the simplest way to plug the geoblocking check of incoming connections.

### ip-auth service

For that purpose a new service 'ip-auth' is introduced. Its main responsibility is to fulfil the [external authorizer](https://istio.io/latest/docs/tasks/security/authorization/authz-custom/) contract:
- listening to the connections from Envoy proxy
- deciding whether to allow/disallow an incoming connection (based on headers and current list of IP ranges)
- responding with HTTP 200 to allow an incoming connection or with HTTP 403 to disallow it

High level overview:

![IP Auth](../../assets/geoblocking.svg)

### Modes of operation

The IP Auth service offers two modes of operation:
1. with IP range allow/block list populated by customer
2. with SAP internal service (only SAP internal customers)

#### IP range allow/block list populated by customer

In this mode the list of blocked IP ranges is read from a config map and stored in ip-auth application memory. The end-user may update the list of IP ranges at any time, so the IP-auth application is obliged to refresh it regularly. 

Example of an IP range allow/block list:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: ip-range-list
data:
  allow: |
    - 192.0.2.10/32
    - 2001:0002:6c::430
  block: |
    - 192.0.2.0/24
    - 198.51.100.0/24
    - 203.0.113.0/24
    - 2001:0002::/48
```

The allow list should take precedence over the block list, which may be useful in allowing the narrower range within blocked broader range. Lists may contain both IPv4 and IPv6 ranges.

![Static list](../../assets/geoblocking-custom-list.svg)

#### Usage of SAP internal service

In this mode the list of blocked IP ranges is received from an SAP internal service. In order to connect to it the ip-auth requires a secret with SAP internal service credentials. The list of blocked IP ranges is then in application memory and additionally in a configmap, which works as a persistent cache. This approach limits the number of list downloads and makes the whole solution more reliable if SAP internal service is not available. The list of IP ranges should be refreshed once per hour.

Additionally, the ip-auth service uses SAP internal service to report the following events:
- policy list consumption (success, failure, unchanged)
- access decision (allow, deny)

The ip-auth service should store the IP range allow/block list in a Config Map for caching / fallback purposes. In this case the Config Map may contain more attributes:
- version information - used in events
- etag - used when checking for updates
- lastUpdateCheckTime - the time when the list update has been performed (successfully) 
- lastUpdateTime - the time when the list has been updated

The above Config Map should only be used by the ip-auth service for caching/fallback purpose, it can't be treated as an official contract. It may be replaced with some other solution at any time without further notice.

Example of an IP range allow/block list cache:

```
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ip-range-list-cache
data:
  version: ip-list-20240912-2200
  etag: PeJwSe7ncNQchtGbkCPJ
  lastUpdateCheckTime: "2024-09-17T15:28:50Z"
  lastUpdateTime: "2024-09-17T14:28:50Z"
  allow: |
    - 192.0.2.10/32
    - 2001:0002:6c::430
  block: |
    - 192.0.2.0/24
    - 198.51.100.0/24
    - 203.0.113.0/24
    - 2001:0002::/48
```

![SAP internal service](../../assets/geoblocking-SAP-service.svg)

### Geoblocking CR and controller

In order to ensure reliability and configurability a new Geoblocking Custom Resource and a new Geoblocking Controller is introduced. The controller is responsible for:
- managing the ip-auth service deployment
- managing authorization policy that plugs ip-auth authorizer to all incoming requests
- performing configuration checks (like external traffic policy)
- reporting geoblocking state

There should be only one Geoblocking resource. The controller should set the additional Geoblocking CR in `Error` state, similarly to [APIGateway CR](https://github.com/kyma-project/api-gateway/blob/05ed57f3299f42c2565ed1b28b84dee808a5213b/controllers/operator/apigateway_controller.go#L100).

![Geoblocking controller](../../assets/geoblocking-cr-controller.svg)

#### CR examples

##### Usage of SAP internal service

```yaml
apiVersion: geoblocking.kyma-project.io/v1alpha1
kind: GeoBlocking
metadata:
  name: geoblocking
  namespace: kyma-system
spec:
  gbService:
    secret: "gb-secret"
    tokenUrl: https://oauth.example.com/oauth2/token
    ipRange:
      url: https://lists.example.com/lists/some-list
      refreshInSeconds: 3600
    events:
      url: https://events.example.com/events
      lobId: "kyma"
      productId: "some-product"
      systemId: "some-system"
```

##### IP range allow/block list populated by customer
```yaml
apiVersion: geoblocking.kyma-project.io/v1alpha1
kind: GeoBlocking
metadata:
    name: geoblocking
    namespace: kyma-system
spec:
  ipRange:
    configMap: "my-own-ip-range-list"
```

### Technical details

#### Policy download optimization

In order to reduce unnecessary updates of the list of blocked IP ranges the ip-auth service should use ETag mechanism, which is supported by the SAP internal service.

The local copy (stored in a Config Map) should be used to optimize the amount of connections to the SAP internal service:
- if the local copy contains a newer version (lastUpdateTime) -> load it
- if the update check has been recently performed (now - lastUpdateCheckTime < refreshInterval) -> don't check it again
- if the update check hasn't been recently performed (now - lastUpdateCheckTime >= refreshInterval) -> check for an update
- if there is a newer version available in the central SAP Geoblocking service -> download, apply it, update a local copy and set lastUpdateCheckTime and lastUpdateTime to the current time
- if there is no newer version -> update lastUpdateCheckTime time (so other pods may skip the check)
- if there is a problem with checking for an update -> log a warning
- if there is a problem with updating the local copy -> log a warning and use a new version anyway

Pods should slightly randomize an update check time to benefit from the above optimization (minimize a risk of performing the check by multiple pods at the same time).

#### Quick IP check

The list of blocked IP ranges may be big (thousands of entries), so analyzing whether an IP address matches any IP range in the list is too time consuming.

The list of IP ranges changes rarely, so it is better to build a data structure that supports quick comparision of IP addresses. The good candidate is a [radix tree](https://en.wikipedia.org/wiki/Radix_tree).

#### Events retention

In order to not cause unnecessary delays the access events should be sent asynchronously:
- after making access decision the ip-auth should generate an event and store it in a memory queue
- separate thread should grab events from the queue and send them to the SAP internal service.

This approach may cause issues if SAP internal service responds slowly or does not respond at all, because events would be consuming memory. System availability is a key factor here, so it is acceptable to drop events in order to ensure stability (prevent out of memory issues).

Events should have some retry number, so the application can drop them if maximum number of retries is exceeded. Events should be also dropped if the queue is full. All such cases should be properly logged.

#### Headers used in check

The ip-auth service should take the following HTTP headers into consideration:
1. x-envoy-external-address - contains a single trusted IP address
2. x-forwarded-for - contains appendable list of IP addresses (modified by each compliant proxy)

IP-auth should block the connection request if any IP address in any of above headers belongs to any IP range in the block list (unless they are also in allow list).

#### ip-auth Deployment settings

The main responsibility of the controller is to reconcile the ip-auth deployment. The deployment details should be hardcoded for simplicity, similarly to [oathkeeper deployment](https://github.com/kyma-project/api-gateway/blob/fdef70be4ca9a9319ae4c547f4f6dad8c73b9846/internal/reconciliations/oathkeeper/deployment.yaml).

This includes at least the following details:
- deployment:
  - replicas / scaling
  - update strategy
  - Docker image
  - resources (limits, requests)
  - probes (liveness, readiness)
  - exposed ports
  - security context (user, etc.)
  - service account
  - configmap mount (ip allow / block list)
  - secret mount (secrets for Geoblocking service)
- service:
  - selectors
  - ports

#### Authorization Policy settings

For simplicity reasons, we assume that geoblocking would be applied globally for the whole application.

This means that it just applies to the whole Istio Ingress Gateway and there is no need expose any additional configuration in this scope.

For now we assume that the feature works with default Istio Ingress Gateway installed by the Istio module.

#### Namespaces

Geoblocking resources would be placed in the following namespaces:

| Resource                           | Who creates it                | Namespace                          |
| ---------------------------------- | ----------------------------- | ---------------------------------- |
| Geoblocking Operator               | Power user                    | kyma-system                        |
| Geoblocking CR                     | Power user                    | kyma-system                        |
| ip-auth Deployment                 | Geoblocking Operator          | kyma-system                        |
| ip-auth Service                    | Geoblocking Operator          | kyma-system                        |
| ip-auth Secret                     | Power user                    | kyma-system                        |
| IP ranges Config Map with input    | ip-auth or External customer  | kyma-system                        |
| IP ranges Config Map with cache    | ip-auth or External customer  | kyma-system                        |
| Authorization Policy               | Geoblocking Operator          | istio-system                       |

![Geoblocking namespaces](../../assets/geoblocking-namespaces.svg)

## Considered alternative architectural approaches

### Workload resources and event queue parameters

Events are stored in a queue, so there is a correlation between:
- queue size
- event TTL
- ip-auth pod memory resources

These parameters may be:
- hardcoded
- calculated based on current memory setup
- configurable in the Custom Resource

Configuration could look like:
```yaml
spec:
  gbService:
    events:
      queue:
        size: 100
        ttl: 360
  deployment: 
    resources:
      requests:
        memory: "64Mi"
        cpu: "250m"
      limits:
        memory: "128Mi"
        cpu: "500m"
```

Decision: it doesn't make sense to expose so many technical parameters, especially that there would be no easy way to determine the above values. It is better to hardcode these parameters with some reasonable values (for now), potentially providing multiple variants (evaluation and production profiles).

### Configurability of the rules in Authorization Policy

Istio allows to specify a lot of conditions that are used to link a request with a given Authorization Policy: https://istio.io/latest/docs/reference/config/security/authorization-policy/#Rule

We can either:
- expose them and make it configurable in the CR
- hide it and doesn't allow such configurability

Decision: it doesn't make sense to expose so many detailed parameters, because the idea behind Geoblocking is to be compliant with regulations, which are usually very global (apply to the whole company / product / application).

### IP range allow/block list fallback/caching

The list of IP ranges is processed by each ip-auth container and stored in its memory and regularly refreshed using geoblocking service. If geoblocking service doesn't work then the ip-auth containers may work as before, the list is not refreshed.

However, the problem is when the ip-auth container needs to be restarted when geoblocking service is not available, so it can't download the list of IP ranges. This may be the case during an upgrade of ip-auth or in the event of runtime problems (for example ouf of memory).

It is not acceptable to allow the traffic as a fallback in such case.

The following options have been considered:

| Option             | Advantages                                              | Disadvantages                                           |
| ------------------ | ------------------------------------------------------- | ------------------------------------------------------- |
| Allow all fallback | Very simple                                             | Not acceptable from security perspective                |
|                    |                                                         |                                                         |
| No fallback        | Very simple                                             | Downtime if geoblocking service doesn't work            |
|                    |                                                         | Every container causes load on geoblocking service      |
|                    |                                                         |                                                         |
| Caching service    | Extensibility (may cache more things in future)         | Complex - another component to take care                |
|                    | Works well with restart / upgrade / scaling             |                                                         |
|                    | No downtime if geoblocking service doesn't work         |                                                         |
|                    | Less load on geoblocking service                        |                                                         |
|                    |                                                         |                                                         |
| Config map         | Still simple                                            | Configmap capacity allows max ~50000 IP ranges          |
|                    | Works well with restart / upgrade / scaling             | IP list integrity not guaranteed                        |
|                    | No downtime if geoblocking service doesn't work         |                                                         |
|                    | Less load on geoblocking service                        |                                                         |
|                    |                                                         |                                                         |
| Ephemeral volume   | Still simple                                            | Every container causes load on geoblocking service      |
|                    | No downtime if geoblocking service doesn't work         | Doesn't help in case of restart / upgrade / scaling     |
|                    |                                                         |                                                         |
| Persistent volume  | No downtime if geoblocking service doesn't work         | Depends on the cloud infrastructure (ReadWriteMany)     |
|                    | Less load on geoblocking service                        |                                                         |
|                    | Works well with restart / upgrade / scaling             |                                                         |

Decision: Let's use a Config Map as a fallback and cache for IP range allow/block list. However, let's split it technically from the Config Map that contains IP ranges provided by the end-user, so:
- the Config Map containing custom IP range allow/block list is configured by the user, it becomes a contract
- the Config Map containing IP range allow/block list downloaded from the SAP internal service is a Kyma internal resource and the implementation may change at any time, so no other module should use it

Consequence: Config map capacity may be exceeded in future, which may require immediate attention.

### ip-auth deployment configurability

The Geoblocking operator main functionality is the deployment of the ip-auth service. The deployment requires multiple parameters (like Docker image, resources). These parameters may be either configurable in the custom resource or just hardcoded.

Decision: Hardcode deployment details in the API Gateway module.

Consequence: It won't be possible to change the deployment details. 

### Propagation of ip-auth service issues

Geoblocking controller is responsible for the reconciliation of the Geoblocking custom resource, in particular: creating deployment of ip-auth service, creation of Authorization Policy, doing some configuration checks, etc. It may report issues related to above actions in the Geoblocking custom resource status.

However, some issues may be visible only in the ip-auth service. The good example is a problem with the communication with SAP internal service (like a wrong secret).

There is no easy way to propagate such issues to the Geoblocking CR. This would require a dedicated mechanism - like a 'health' endpoint exposed by ip-auth containers, so it can be analyzed by the Geoblocking controller.

Decision: Don't implement additional mechanisms and follow the 'standard' approach, so the controller reports issues in the CR only from its own layer, while the ip-auth service reports issues in its log or via liveness/readiness probes (like every other workload).

Consequence: Geoblocking CR won't be responsible for presenting Geoblocking 'health'. Support or developers must be aware that they should also inspect ip-auth logs in case of issues.

### Namespaces

Geoblocking mechanism touches multiple resources in multiple areas:
- Istio configuration
- Ingress gateway
- Gateways in general
- Authorization Policy
- Authorizer workload (service, deployment)
- Configuration of SAP internal service
- Custom IP range allow/block list
- IP range allow/block list cache

It is a mix-in of different scopes - those managed by Kyma, by Istio and by the end-users.

There are multiple factors that need to be taken into consideration:
- Permissions - the Operator needs more permissions to access multiple namespaces
- Kyma resources (workloads) should run in kyma-system namespace
- kyma-system namespace should not be used by the end-users
- Ingress Gateway is created by default in the istio-system namespace
- Users may influence Istio (via Gateway API) to create Ingress Gateway in different namespace
- Authorization Policy references Ingress Gateway via selector, so they should be in the same namespace
- Gateway CR is supposed to be a singleton (for now)

There are multiple approaches possible:
- Geoblocking CR being close to the gateway (by default istio-system)
- Geoblocking CR being in end-user's (workload) namespace
- Geoblocking CR being in kyma-namespace (close to ip-auth deployment)
- ip-auth deployment being run in user's namespace (allows multiple user configurations)
- ip-auth deployment being run in kyma-system namespace (along with other Kyma modules)
- all resources in istio-system (avoids multiple namespaces, but breaks Istio separation)

Decision: We can't avoid working with multiple namespaces. Taking into consideration the fact that Geoblocking (being compliance requirement) is rather an application-global functionality, it makes sense to have a single Geoblocking CR (implying single configuration). Thus, it seems that the optimal solution is to keep all resources in kyma-system namespace, expect for the Authorization Policy that references Ingress Gateway (being in istio-system by default).

Consequences: Power user would have to configure Geoblocking in kyma-system (create CR, provide a Secret).

### Istio ip-auth authorizer declaration

The ip-auth service needs to be declared as custom authorizer in the Istio configuration.

There are several possibilities to do it:
- Geoblocking controller may modify Istio configuration (breaks separation)
- Geoblocking controller may modify the Istio CR (still breaks the separation)
- End-user may do it
- It may be always configured by the Istio module

Decision: End users should do it, because they need to modify it anyway (External Traffic Policy). The Geoblocking Operator should check whether it is properly configured (Authorizer declaration, External Traffic Policy set properly, etc.)

Consequence: Geoblocking won't work OOTB after applying Geoblocking CR
