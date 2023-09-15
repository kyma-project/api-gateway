# Allowed domains in API Gateway Controller

API Gateway Controller uses an allowlist of domains for which it allows to expose Services. Every time a user creates a new APIRule custom resource (CR) for a Service, API Gateway Controller checks the domain of the Service specified in the CR against the allowlist. If it matches an allowed entry, API Gateway Controller creates a VirtualService and Oathkeeper Access Rules for the Service according to the details specified in the CR. If the domain is not allowed, the Controller creates neither a VirtualService nor Oathkeeper Access Rules and, as a result, does not expose the Service.

If the domain does not match the allowlist, API Gateway Controller sets an appropriate validation status on the APIRule CR created for the Service.

By default, the only allowed domain is the domain of the Kyma cluster.
