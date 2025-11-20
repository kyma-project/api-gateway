export default [
  { text: 'Quick Start Guide', link: './os-quick-start.md' },
  { text: 'Tutorials', link: './tutorials/README', collapsed: true, items: [
    { text: 'Set Up a Custom Domain', link: './tutorials/01-10-setup-custom-domain-for-workload.md' },
    { text: 'Set Up a TLS Gateway', link: './tutorials/01-20-set-up-tls-gateway.md' },
    { text: 'Mutual TLS Authentication', link: './tutorials/01-10-mtls-authentication/README.md', collapsed: true, items: [
      { text: 'Configure mTLS Authentication in SAP BTP, Kyma Runtime', link: './tutorials/01-30-mutual-tls-authentication/configure-mtls-gardener-certs.md' },
      { text: 'Configure mTLS Authentication on k3d', link: './tutorials/01-30-mutual-tls-authentication/configure-mtls-k3d.md' },
    ]},
    { text: 'Expose a Workload', link: './tutorials/01-40-expose-workload/README.md', collapsed: true, items: [
      { text: 'Expose a Workload', link: './tutorials/01-40-expose-workload/01-40-expose-workload-apigateway.md' },
      { text: 'Expose Multiple Workloads', link: './tutorials/01-40-expose-workload/01-41-expose-multiple-workloads.md' },
      { text: 'Expose Workloads in Multiple Namespaces', link: './tutorials/01-40-expose-workload/01-42-expose-workloads-multiple-namespaces.md' },
      { text: 'Use a Short Host', link: './tutorials/01-40-expose-workload/01-43-expose-workload-short-host-name.md' }
    ]},
    { text: 'Expose and Secure a Workload', link: './tutorials/01-50-expose-and-secure-a-workload/README.md', collapsed: true, items: [
      { text: 'Get JSON Web Tokens (JWT)', link: './tutorials/01-50-expose-and-secure-a-workload/01-51-get-jwt.md' },
      { text: 'Secure a Workload with JWT', link: './tutorials/01-50-expose-and-secure-a-workload/01-52-expose-and-secure-workload-jwt.md' },
      { text: 'Secure a Workload with extAuth', link: './tutorials/01-50-expose-and-secure-a-workload/01-53-expose-and-secure-workload-ext-auth.md' },
      { text: 'Secure a Workload with a Certificate', link: './tutorials/01-50-expose-and-secure-a-workload/01-54-expose-and-secure-workload-with-certificate.md' },
      { text: 'Configure IP-Based Access with XFF', link: './tutorials/01-50-expose-and-secure-a-workload/01-55-ip-based-access-with-xff.md' }
    ]},
    { text: 'Security', link: './tutorials/01-60-security/README.md', collapsed: true, items: [
      { text: 'Prepare Self-Signed Root CA and Client Certificates', link: './tutorials/01-60-security/01-61-mtls-selfsign-client-certicate.md' },
      { text: 'Set Up a Custom Identity Provider', link: './tutorials/01-60-security/01-62-set-up-idp.md' },
    ]},
    { text: 'Configuring Local Rate Limiting', link: './tutorials/01-70-local-rate-limit.md' },
  ]},
  { text: 'Custom Resources', link: './custom-resources/README.md', collapsed: true, items: [
    { text: 'APIGateway Custom Resource', link: './custom-resources/apigateway/README.md', collapsed: true, items: [
      { text: 'Specification', link: './custom-resources/apigateway/04-00-apigateway-custom-resource.md' },
      { text: 'Kyma Gateway', link: './custom-resources/apigateway/04-10-kyma-gateway.md' },
      { text: 'Oathkeeper Dependency', link: './custom-resources/apigateway/04-20-oathkeeper.md' }
    ]},
    { text: 'APIRule Custom Resource', link: './custom-resources/apirule/README.md', collapsed: true, items: [
        { text: 'APIRule v2 Custom Resource', link: './custom-resources/apirule/04-10-apirule-custom-resource.md' },
        { text: 'APIRule Access Strategies', link: './custom-resources/apirule/04-15-api-rule-access-strategies.md' },
        { text: 'Ordering Rules in APIRule v2', link: './custom-resources/apirule/04-20-significance-of-rule-path-and-method-order.md' }
    ]},
    { text: 'RateLimit Custom Resource', link: './custom-resources/ratelimit/04-00-ratelimit.md' },
  ]},
  { text: 'APIRule Migration', link: './apirule-migration/README.md', collapsed: true, items: [
    { text: 'Retrieve v1beta1 spec', link: './apirule-migration/01-81-retrieve-v1beta1-spec.md' },
    { text: 'Migrate Multiple APIRules Targeting the Same Workload', link: './apirule-migration/01-90-migrate-multiple-apirules-targeting-same-workload.md' },
    { text: 'Migrate jwt Handlers', link: './apirule-migration/01-83-migrate-jwt-v1beta1-to-v2.md' },
    { text: 'Migrate noop, no_auth, allow Handlers', link: './apirule-migration/01-82-migrate-allow-noop-no_auth-v1beta1-to-v2.md' },
    { text: 'Migrate Ory-based Handlers', link: './apirule-migration/01-84-migrate-oauth2-v1beta1-to-v2.md' },
    { text: 'Changes in APIRule v2', link: './custom-resources/apirule/04-70-changes-in-apirule-v2.md' },
    { text: 'FAQ', link: './apirule-migration/01-85-apirule-migration-faq.md' }
  ]},
  { text: 'Technical Reference', link: './technical-reference/README.md', collapsed: true, items: [
    { text: 'Kyma API Gateway Operator Parameters', link: './technical-reference/05-00-api-gateway-operator-parameters.md' },
    { text: 'Ory Limitations', link: './technical-reference/05-50-ory-limitations.md' },
    { text: 'Api Gateway Request Flow', link: './technical-reference/05-10-api-gateway-request-flow.md' }
  ]},
  { text: 'Troubleshooting Guides', link: './troubleshooting-guides/README.md', collapsed: true, items: [
    { text: 'Issues When Creating APIRule', link: './troubleshooting-guides/03-40-api-rule-troubleshooting.md' },
    { text: 'Blocked In-Cluster Communication', link: './troubleshooting-guides/03-80-blocked-in-cluster-communication.md' },
    { text: 'Missing Rules in APIRule `v2alpha1`', link: './troubleshooting-guides/03-84-missing-rules-apirule-v2alpha1.md' },
    { text: 'Missing Rules in APIRule `v2`', link: './troubleshooting-guides/03-81-missing-rules-apirule-v2.md' },
    { text: 'APIRules Missing from Kyma Dashboard', link: './troubleshooting-guides/03-83-apirules-are-missing-from-kyma-dashboard.md' },
    { text: 'Changed Status Schema in APIRule `v2`', link: './troubleshooting-guides/03-82-changed-status-schema-apirule-v2.md' },
    { text: 'APIRule Conversion Webhook Failed', link: './troubleshooting-guides/03-60-apirule-conversion-webhook-cert-issue.md' },
    { text: 'Certificate Management: Issuer Not Created', link: './troubleshooting-guides/03-20-cert-mgt-issuer-not-created.md' },
    { text: 'Kyma Gateway Not Reachable', link: './troubleshooting-guides/03-30-gateway-not-reachable.md' },
    { text: '401 Unauthorized or 403 Forbidden', link: './troubleshooting-guides/03-01-401-unauthorized-403-forbidden.md' },
    { text: 'Connection Refused or Timeout', link: './troubleshooting-guides/03-10-dns-mgt-connection-refused.md' },
    { text: 'Could not Resolve Host', link: './troubleshooting-guides/03-11-dns-mgt-could-not-resolve-host.md' },
    { text: 'Resource Ignored by the Controller', link: './troubleshooting-guides/03-12-dns-mgt-resource-ignored.md' },
    { text: 'Issues with Gardener Certificates', link: './troubleshooting-guides/03-50-certificates-gardener.md' },
  ]}
]
