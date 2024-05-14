# Certificate management for APIRule conversion webhook with a new controller

## Status

Proposed

## Context

We introduce a conversion webhook for converting `v1beta1` from and to `v1beta2` APIRule versions. For it we need a valid x509 certificate needed for the webhook server. We must provide implementation of certificate handling such as creating, verifying and renewal. Previously for conversion webhook we did use `CronJob` which scheduled `Job` periodically for handling certificate verification and renewal. This required new image to be maintained, build and released.

## Decision

1. We like to handle this creation, verifying and renewal of the required certificate integrated into our Module operator with a new Kubernetes controller.
2. We do not create additional image that would handle certificate mangement with `CronJob` which showed quite a few limitations (`Job` updating for instance) and inconviniences in regards of observability.
3. New controller will reconcile predefined `Secret` named `api-gateway-webhook-certificate` in `kyma-system` namespace that is holding the data for the Certificate.
4. We introduce an Kubernetes `init container` to the operator deployment which will handle initial creation of the predefined Secret holding the Certificate.
5. Newly created `Secret` will have `gateways.operator.kyma-project.io/certificate` finalizer to prevent accidental deletion.
6. We delegate function to the Webhoook server for obtaining current Certificate which will fully automate Certificate renewal process.
7. We rotate the Certificate 14 days before expiration and we create Certificate with 90 days validity. SAP recommendation for SSL server certificates is one year validity.

## Consequences

APIRule conversion works out of the box with integrated conversion webhook started with the controller manager and certificate is managed by us in our Module operator falling Kubernetes controller reconciling pattern.
