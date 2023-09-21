# Blocked Services in APIGateway Controller

APIGateway Controller uses a blocklist of Services for which it does not create either a VirtualService or Oathkeeper Access Rules. As a result, these Services cannot be exposed. Every time a user creates a new APIRule custom resource (CR) for a Service, APIGateway Controller checks the name of the Service specified in the CR against the blocklist. If it matches a blocklisted entry, APIGateway Controller sets an appropriate validation status on the APIRule CR created for that Service.

The blocklist works as a security measure and prevents users from exposing vital internal Services of Kubernetes and Istio.
