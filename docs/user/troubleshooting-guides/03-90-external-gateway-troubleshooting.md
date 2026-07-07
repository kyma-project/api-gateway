# ExternalGateway Troubleshooting

If your ExternalGateway or the traffic that flows through it is not behaving as expected, start with:

```bash
kubectl describe externalgateway <name> -n <namespace>
```

If `Ready=False`, jump to [ExternalGateway readiness conditions](#externalgateway-readiness-conditions).

Otherwise, match your symptom to one of the six failure modes below.

## 1. Request doesn't reach Kyma

**Symptom.** You send a request to your ExternalGateway host, but the connection is terminated without any Envoy response headers.

**Cause.** The request never reaches the Kyma Istio ingress gateway. Common root causes:

- The DNSEntry for the internal domain has not been provisioned yet, or its target LoadBalancer address is empty (Gardener has not populated `istio-ingressgateway`).
- The external gateway (the appliance outside the cluster) is not configured to route to the Kyma internal domain.
- The Kyma cluster's LoadBalancer Service is not reachable from where the external gateway resolves it.

**Solution.**

1. Check the ExternalGateway status first. If `Ready=False`, see [Readiness conditions](#externalgateway-readiness-conditions).

2. Verify the DNSEntry for the internal domain exists and is Ready:

   ```bash
   kubectl get dnsentry -n istio-system | grep <externalgateway-name>
   kubectl describe dnsentry -n istio-system <dnsentry-name>
   ```

3. Verify the Istio ingress gateway Service has a LoadBalancer address:

   ```bash
   kubectl get svc istio-ingressgateway -n istio-system
   ```

   If the `EXTERNAL-IP` column is empty or `<pending>`, the request cannot reach the cluster.

## 2. TLS handshake rejected

**Symptom.** Your request reaches the Kyma Istio ingress gateway, but the TLS handshake fails with one of:

- `tls: unknown certificate authority`
- `alert: bad certificate`
- `alert: unknown ca`
- The client certificate is rejected.

Common root causes:

**Cause.**

- The external gateway's client certificate is not signed by the CA in the referenced `spec.caSecretRef` Secret.
- The `spec.caSecretRef` Secret is missing, empty, or does not contain a valid PEM-encoded CA.
- The Istio Gateway is configured for mTLS `MUTUAL`, but the external gateway is not sending a client certificate.
- The SNI in the connection does not match either the external or the internal domain configured on the Istio Gateway.

**Solution.**

1. Check the ExternalGateway status. Look for `Ready=False` with reason `CASecretNotFound`, `CASecretKeyAmbiguous`, or `CASecretInvalid`. See [readiness conditions](#externalgateway-readiness-conditions).

2. Inspect the CA Secret content:

   ```bash
   kubectl get secret <ca-secret-name> -n <namespace> -o jsonpath='{.data.ca\.crt}' | base64 -d | openssl x509 -noout -subject -issuer -dates
   ```

   Confirm it is a valid, unexpired CA certificate.

## 3. Certificate validation filter failed

**Symptom.** TLS handshake completes successfully, but the request is answered with HTTP `403 Forbidden` and a body of `Forbidden` (no JSON envelope). Requests never reach your workload.

**Cause.** The Kyma-side certificate validation EnvoyFilter rejected the external gateway's client certificate because its **Subject** does not match any entry in the RegionsConfigMap for the configured region.

Common root causes:

- `spec.region` on the ExternalGateway points at the wrong region entry.

**Solution.**

1. Check the ExternalGateway status. Look for `Ready=False` with reason `RegionNotFound`, `RegionHasNoSubjects`, or `RegionSubjectMalformed`. See [readiness conditions](#externalgateway-readiness-conditions).

2. Print the RegionsConfigMap and confirm your region has the expected subjects:

   ```bash
   kubectl get configmap <regions-configmap> -n <namespace> -o yaml
   ```

## 4. Host not routed

**Symptom.** TLS and certificate validation both succeed, but you receive HTTP `404 Not Found` or `502 Bad Gateway` with an Envoy-shaped body. The request never reaches your workload.

**Cause.** The `Host` header on the request does not match any VirtualService or APIRule `hosts`.

- Your APIRule `spec.hosts` does not include the domain in the `Host` header.
- The external gateway sends a different `Host` header than you expect (for example, a `.internal` variant when the APIRule expects the customer-facing domain).
- The APIRule references the wrong ExternalGateway via `spec.externalGateway: ns/name`.

**Solution.**

1. Verify the APIRule and its host list:

   ```bash
   kubectl get apirule <name> -n <namespace> -o yaml
   ```

   Confirm `spec.hosts` contains the exact domain you send in the `Host` header, and that `spec.externalGateway` matches the ExternalGateway name and namespace.

2. Inspect the corresponding VirtualService:

   ```bash
   kubectl get virtualservice -n <namespace> -o yaml
   ```

   Confirm `spec.hosts` matches.

## 5. Workload unreachable

**Symptom.** Routing succeeds but the response is HTTP `503 Service Unavailable`, `504 Gateway Timeout`, or an Envoy `no_healthy_upstream` body.

**Cause.**

- The APIRule references a Service that does not exist in the target namespace.
- The Service exists but has no Endpoints (no matching Pods, or the Pods are not Ready).
- The Service port does not match the workload's container port.
- Istio sidecar injection is disabled on the workload namespace and the workload does not participate in the mesh.

**Solution.**

1. Check the target Service:

   ```bash
   kubectl get svc <service-name> -n <namespace>
   ```

2. Check the Pods behind the Service:

   ```bash
   kubectl get pods -n <namespace> -l <selector-from-svc>
   ```

   The target Pods should be in `Running` state.

3. Confirm the port in the APIRule matches the Service port:

   ```bash
   kubectl get apirule <name> -n <namespace> -o jsonpath='{.spec.service.port}'
   kubectl get svc <service-name> -n <namespace> -o jsonpath='{.spec.ports[*].port}'
   ```

4. Confirm sidecar injection on the namespace:

   ```bash
   kubectl get namespace <namespace> -o jsonpath='{.metadata.labels}'
   ```

   Look for `istio-injection=enabled`.

## ExternalGateway readiness conditions

Run `kubectl describe externalgateway <name> -n <namespace>` and look at the `Ready` condition. Match its `reason` field against this table.

| Reason                         | Meaning                                                                                                                                 |
|--------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| `InternalDomainTooLong`        | The assembled internal domain (`{kymaSubdomain}.{gateway-domain}`) is longer than 64 characters — the X.509 CommonName limit.           |
| `CASecretNotFound`             | The Secret referenced by `spec.caSecretRef` does not exist in the referenced namespace (defaults to the ExternalGateway's namespace).   |
| `CASecretKeyAmbiguous`         | The referenced Secret has more than one data key and none of them is `ca.crt`.                                                          |
| `CASecretInvalid`              | The Secret exists and the key is present, but the value is empty.                                                                       |
| `RegionsConfigMapNotFound`     | The ConfigMap named by `spec.regionsConfigMap` does not exist in the ExternalGateway's namespace.                                       |
| `RegionsConfigMapKeyAmbiguous` | The ConfigMap has more than one data key and none of them is `regions.yaml`.                                                            |
| `RegionsConfigMapInvalid`      | The ConfigMap value does not parse as YAML, or the parsed structure has zero regions.                                                   |
| `RegionNotFound`               | `spec.region` does not match any region entry in the ConfigMap.                                                                         |
| `RegionHasNoSubjects`          | The region exists but its `subjects` list is empty.                                                                                     |
| `ExternalDomainConflict`       | Another ExternalGateway in the cluster already uses the same `spec.externalDomain`. The message names the conflicting `namespace/name`. |
| `ReconciliationFailed`         | Fallback for any error not classified above.                                                                                            |
