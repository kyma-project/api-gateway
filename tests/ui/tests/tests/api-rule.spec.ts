import 'cypress-file-upload';
import { generateNamespaceName, generateRandomName } from "../support";

const apiRuleDefaultPath = "/.*";

context("API Rule", () => {

    let apiRuleName = "";
    let namespaceName = "";
    let serviceName = "";

    beforeEach(() => {
        apiRuleName = generateRandomName("test-api-rule");
        namespaceName = generateNamespaceName();
        serviceName = generateRandomName("test-service");
        cy.loginAndSelectCluster();
        cy.createNamespace(namespaceName);
        cy.createService(serviceName, namespaceName);
    });

    afterEach(() => {
        cy.deleteNamespace(namespaceName);
    });

    it("should create no_auth APIRule as default", () => {
        cy.navigateToApiRuleList(namespaceName);

        cy.clickCreateButton();

        cy.apiRuleTypeName(apiRuleName);
        cy.apiRuleSelectService(serviceName);
        cy.apiRuleTypeServicePort("80");
        cy.apiRuleTypeHost(apiRuleName);

        cy.clickCreateButton();

        cy.hasStatusLabel("OK");
        cy.contains(apiRuleDefaultPath).should('exist');
        cy.contains('Rules #1', { timeout: 10000 }).click();
        cy.contains('no_auth').should('exist');
    });

    it("should create oauth2_introspection APIRule", () => {
        cy.navigateToApiRuleList(namespaceName);

        cy.clickCreateButton();

        cy.apiRuleTypeName(apiRuleName);
        cy.apiRuleSelectService(serviceName);
        cy.apiRuleTypeServicePort("80");
        cy.apiRuleTypeHost(apiRuleName);

        cy.apiRuleSelectAccessStrategy("oauth2_introspection");
        cy.get('[aria-label="expand Required Scope"]', { log: false, }).click();
        cy.inputClearAndType('[data-testid="spec.rules.0.accessStrategies.0.config.required_scope.0"]:visible', "read");

        cy.apiRuleSelectMethod("POST")

        cy.clickCreateButton();

        // Verify created API Rule
        cy.hasStatusLabel("OK");
        cy.contains(apiRuleDefaultPath).should('exist');
        cy.contains('Rules #1', { timeout: 10000 }).click();
        cy.contains('oauth2_introspection').should('exist');
        cy.contains('read').should('exist');
    });

    it("should create jwt APIRule", () => {
        cy.navigateToApiRuleList(namespaceName);

        cy.clickCreateButton();

        cy.apiRuleTypeName(apiRuleName);
        cy.apiRuleSelectService(serviceName);
        cy.apiRuleTypeServicePort("80");
        cy.apiRuleTypeHost(apiRuleName);
        cy.apiRuleSelectAccessStrategy("jwt");

        cy.apiRuleTypeJwksUrl("https://urls.com");
        cy.contains('JWKS URL: HTTP protocol is not secure, consider using HTTPS',).should('not.exist');

        cy.apiRuleTypeTrustedIssuer("https://trusted.com")
        cy.contains('Trusted Issuers: HTTP protocol is not secure, consider using HTTPS').should('not.exist');

        cy.clickCreateButton();

        // Verify created API Rule
        cy.hasStatusLabel("OK");
        cy.contains(apiRuleDefaultPath).should('exist');
        cy.contains('Rules #1', { timeout: 10000 }).click();
        cy.contains('jwt').should('exist');
        cy.contains('https://urls.com').should('exist');
        cy.contains('https://trusted.com').should('exist');
        cy.contains('Disabling custom CORS Policy is not recommended. Consider setting up CORS yourself').should('exist');
    });

    it('should update oauth2_introspection API Rule to jwt', () => {
        const updatedApiRulePath = "/test-path";

        cy.createApiRule({
            name: apiRuleName,
            namespace: namespaceName,
            service: serviceName,
            host: apiRuleName,
            handler: "oauth2_introspection",
            config: {
                required_scope: ["read"]
            }
        });

        cy.navigateToApiRule(apiRuleName, namespaceName);
        cy.clickEditTab()
        cy.contains(apiRuleName);

        cy.apiRuleTypeRulePath(updatedApiRulePath);

        cy.get('[aria-label="expand Access Strategies"]:visible', { log: false }).first();
        cy.apiRuleSelectAccessStrategy("jwt");

        cy.apiRuleTypeJwksUrl("https://urls.com");
        cy.apiRuleTypeTrustedIssuer("https://trusted.com");

        cy.clickSaveButton();
        cy.clickViewTab();

        // Validate edited API Rule
        cy.hasStatusLabel("OK");
        cy.contains(apiRuleDefaultPath).should('exist');
        cy.contains('Rules #1', { timeout: 20000 }).click();
        cy.contains(updatedApiRulePath).should('exist');

        cy.contains('oauth2_introspection').should('not.exist');
        cy.contains('jwt').should('exist');
        cy.contains('https://urls.com').should('exist');
        cy.contains('https://trusted.com').should('exist');
    });

    it("should display multiple rules are in the list", () => {
        cy.createApiRule({
            name: apiRuleName,
            namespace: namespaceName,
            service: serviceName,
            host: apiRuleName,
            handler: "no_auth"
        });

        const secondApiRuleName = `${apiRuleName}-second`
        cy.createApiRule({
            name: secondApiRuleName,
            namespace: namespaceName,
            service: serviceName,
            host: secondApiRuleName,
            handler: "no_auth"
        });

        cy.navigateToApiRuleList(namespaceName);
        cy.hasTableRowNumberContaining(1, apiRuleName);
        cy.hasTableRowNumberContaining(2, secondApiRuleName);
    });

    it('should update CORS policy', () => {

        cy.createApiRule({
            name: apiRuleName,
            namespace: namespaceName,
            service: serviceName,
            host: apiRuleName,
            handler: "jwt",
            config: {
                jwks_urls: ["https://urls.com"],
                trusted_issuers: ["https://trusted.com"]
            }
        });

        cy.navigateToApiRuleList(namespaceName);

        cy.clickGenericListLink(apiRuleName);
        cy.clickEditTab();

        cy.get('ui5-switch[data-testid="$useCorsPolicy"]')
            .find('[role="switch"]')
            .click();

        cy.get('[aria-label="expand CORS Policy"]')
            .should('be.visible');

        // CORS allow methods
        cy.contains('CORS Allow Methods').should('be.visible');
        cy.get('[data-testid="spec.corsPolicy.allowMethods.GET"]:visible').click();

        // CORS allow origins
        cy.get('[aria-label="expand CORS Allow Origins"]').should('be.visible').contains("Add")

        // CORS allow headers
        cy.get('[aria-label="expand CORS Allow Headers"]').should('be.visible').click();
        cy.inputClearAndType('[data-testid="spec.corsPolicy.allowHeaders.0"]', "Allowed-Header");

        // CORS allow headers
        cy.get('[aria-label="expand CORS Expose Headers"]').should('be.visible').click();
        cy.inputClearAndType('[data-testid="spec.corsPolicy.exposeHeaders.0"]', "Exposed-Header");

        // CORS allow credentials
        cy.get('ui5-switch[data-testid="spec.corsPolicy.allowCredentials"]')
            .find('[role="switch"]')
            .click();

        // CORS Max Age
        cy.inputClearAndType('[data-testid="spec.corsPolicy.maxAge"]', "10s");

        cy.clickSaveButton();
        cy.clickViewTab();

        // Verify CORS policy
        cy.hasStatusLabel("OK");
        cy.contains(apiRuleDefaultPath).should('exist');

        cy.contains('CORS Allow Methods').should('exist').parent().contains('GET').should('exist');
        cy.contains('CORS Expose Headers').should('exist').parent().contains('Exposed-Header').should('exist')
        cy.contains('CORS Allow Headers').should('exist').parent().contains('Allowed-Header').should('exist')
        cy.contains('CORS Allow Credentials').should('exist').parent().contains('true').should('exist')
        cy.contains('CORS Max Age').should('exist').parent().contains('10s').should('exist')
    });

    it('should show alert warning if Gateway is in wrong format', () => {
        cy.createApiRule({
            name: apiRuleName,
            namespace: namespaceName,
            service: serviceName,
            host: apiRuleName,
            handler: "no_auth",
            gateway: "kyma-system"
        });

        cy.navigateToApiRuleList(namespaceName);
        cy.clickGenericListLink(apiRuleName);
        cy.contains('Gateway must be in the format \'{NAMESPACE}/{NAME}\'').should('exist')

        cy.clickEditTab();
        cy.contains('Gateway must exist, specify both Namespace and Name').should('exist')

        cy.inputClearAndType('ui5-combobox[placeholder="Select name"]', "kyma-gateway");

        cy.clickSaveButton();
        cy.clickViewTab();

        cy.contains('Gateway must be in the format \'{NAMESPACE}/{NAME}\'').should('not.exist')

        cy.clickEditTab();

        cy.contains('Gateway must exist, specify both Namespace and Name').should('not.exist')
    });

    context("Host", () => {
        context("when APIRule is in OK state", () => {
            it('should build correct link in details view', () => {
                cy.createApiRule({
                    name: apiRuleName,
                    namespace: namespaceName,
                    service: serviceName,
                    host: apiRuleName,
                    handler: "no_auth"
                });

                cy.navigateToApiRule(apiRuleName, namespaceName);
                cy.apiRuleMetadataContainsHostUrl(`https://${apiRuleName}.local.kyma.dev`);
            });
        })

        context("when APIRule is not in OK state", () => {
            it('should have dummy link in details view', () => {
                cy.createApiRule({
                    name: apiRuleName,
                    namespace: namespaceName,
                    service: "not_existent",
                    host: apiRuleName,
                    handler: "no_auth"
                });

                cy.navigateToApiRule(apiRuleName, namespaceName);
                cy.apiRuleMetadataContainsHostUrl(`https://${apiRuleName}`);
            });
        })
    });
});
