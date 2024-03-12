import 'cypress-file-upload';
import {generateNamespaceName, generateRandomName} from "../../support";

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

        // Name
        cy.get('ui5-input[aria-label="APIRule name"]')
            .find('input')
            .type(apiRuleName, {force: true});

        // Service
        cy.chooseComboboxOption('[data-testid="spec.service.name"]', serviceName);

        cy.get('[data-testid="spec.service.port"]:visible', {log: false})
            .find('input')
            .click()
            .clear()
            .type("80");

        // Host
        cy.get('[data-testid="spec.host"]:visible', {log: false})
            .find('input')
            .click()
            .type(apiRuleName);

        cy.get('ui5-dialog')
            .contains('ui5-button', 'Create')
            .should('be.visible')
            .click();

        // Verify created API Rule
        cy.contains(apiRuleName).click();
        cy.contains(apiRuleDefaultPath).should('exist');
        cy.contains('Rules #1', {timeout: 10000}).click();
        cy.contains('no_auth').should('exist');
    });

    it("should create oauth2_introspection APIRule", () => {
        cy.navigateToApiRuleList(namespaceName);

        cy.clickCreateButton();

        // Name
        cy.get('ui5-input[aria-label="APIRule name"]')
            .find('input')
            .type(apiRuleName, {force: true});

        // Service
        cy.chooseComboboxOption('[data-testid="spec.service.name"]', serviceName);

        cy.get('[data-testid="spec.service.port"]:visible', {log: false})
            .find('input')
            .click()
            .clear()
            .type("80");

        // Host
        cy.get('[data-testid="spec.host"]:visible', {log: false})
            .find('input')
            .click()
            .type(apiRuleName);

        // Rules

        // > Access Strategies
        cy.get(
            `ui5-combobox[data-testid="spec.rules.0.accessStrategies.0.handler"]:visible`,
        )
            .find('input')
            .click()
            .clear()
            .type('oauth2_introspection', {force: true});

        cy.get('ui5-li:visible')
            .contains('oauth2_introspection')
            .find('li')
            .click({force: true});

        cy.get('[aria-label="expand Required Scope"]', {
            log: false,
        }).click();

        cy.get(
            '[data-testid="spec.rules.0.accessStrategies.0.config.required_scope.0"]:visible',
        )
            .find('input')
            .click()
            .clear()
            .type('read');

        // > Methods

        cy.get('[data-testid="spec.rules.0.methods.POST"]')

        cy.get(`ui5-checkbox[text="POST"]:visible`).click();

        cy.get('ui5-dialog')
            .contains('ui5-button', 'Create')
            .should('be.visible')
            .click();

        // Verify created API Rule
        cy.contains(apiRuleName).click();
        cy.contains(apiRuleDefaultPath).should('exist');
        cy.contains('Rules #1', {timeout: 10000}).click();
        cy.contains('oauth2_introspection').should('exist');
        cy.contains('read').should('exist');
    });

    it("should create jwt APIRule", () => {
        cy.navigateToApiRuleList(namespaceName);

        cy.clickCreateButton();

        // Name
        cy.get('ui5-input[aria-label="APIRule name"]')
            .find('input')
            .type(apiRuleName, {force: true});

        // Service
        cy.chooseComboboxOption('[data-testid="spec.service.name"]', serviceName);

        cy.get('[data-testid="spec.service.port"]:visible', {log: false})
            .find('input')
            .click()
            .clear()
            .type("80");

        // Host
        cy.get('[data-testid="spec.host"]:visible', {log: false})
            .find('input')
            .click()
            .type(apiRuleName);

        // Rules

        // > Access Strategies
        cy.get(
            `ui5-combobox[data-testid="spec.rules.0.accessStrategies.0.handler"]:visible`,
        )
            .find('input')
            .click()
            .clear()
            .type('jwt', {force: true});

        cy.get('ui5-li:visible')
            .contains('jwt')
            .find('li')
            .click({force: true});

        cy.get('[aria-label="expand JWKS URLs"]:visible', {log: false}).click();

        cy.get(
            '[data-testid="spec.rules.0.accessStrategies.0.config.jwks_urls.0"]:visible',
        )
            .find('input')
            .click()
            .clear()
            .type('https://urls.com');

        cy.contains(
            'JWKS URL: HTTP protocol is not secure, consider using HTTPS',
        ).should('not.exist');

        cy.get('[aria-label="expand JWKS URLs"]:visible', {log: false}).click();

        cy.get('[aria-label="expand Trusted Issuers"]:visible', {
            log: false,
        }).click();

        cy.get(
            '[data-testid="spec.rules.0.accessStrategies.0.config.trusted_issuers.0"]:visible',
        )
            .find('input')
            .click()
            .clear()
            .type('https://trusted.com');

        cy.contains(
            'Trusted Issuers: HTTP protocol is not secure, consider using HTTPS',
        ).should('not.exist');

        cy.get('[aria-label="expand Trusted Issuers"]:visible', {
            log: false,
        }).click();

        cy.get('ui5-dialog')
            .contains('ui5-button', 'Create')
            .should('be.visible')
            .click();

        // Verify created API Rule
        cy.contains(apiRuleName).click();
        cy.contains(apiRuleDefaultPath).should('exist');
        cy.contains('Rules #1', {timeout: 10000}).click();
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

        cy.contains('ui5-button', 'Edit').click();
        cy.contains(apiRuleName);

        cy.get('[data-testid="spec.rules.0.path"]:visible')
            .find('input')
            .click()
            .clear()
            .type(updatedApiRulePath);

        // > Access Strategies
        cy.get('[aria-label="expand Access Strategies"]:visible', {log: false})
            .first();

        cy.get(
            `ui5-combobox[data-testid="spec.rules.0.accessStrategies.0.handler"]:visible`,
        )
            .find('input')
            .click()
            .clear()
            .type('jwt', {force: true});

        cy.get('ui5-li:visible')
            .contains('jwt')
            .find('li')
            .click({force: true});

        cy.get('[aria-label="expand JWKS URLs"]:visible', {log: false}).click();

        cy.get(
            '[data-testid="spec.rules.0.accessStrategies.0.config.jwks_urls.0"]:visible',
        )
            .find('input')
            .click()
            .clear()
            .type('https://urls.com');

        cy.get('[aria-label="expand JWKS URLs"]:visible', {log: false}).click();

        cy.get('[aria-label="expand Trusted Issuers"]:visible', {
            log: false,
        }).click();

        cy.get(
            '[data-testid="spec.rules.0.accessStrategies.0.config.trusted_issuers.0"]:visible',
        )
            .find('input')
            .click()
            .clear()
            .type('https://trusted.com');

        // > Methods
        cy.get('ui5-dialog')
            .contains('ui5-button', 'Update')
            .should('be.visible')
            .click();

        // Validate edited API Rule
        cy.contains(apiRuleName).click();
        cy.contains(apiRuleDefaultPath).should('exist');
        cy.contains('Rules #1', {timeout: 20000}).click();
        cy.contains(updatedApiRulePath).should('exist');

        cy.contains('oauth2_introspection').should('not.exist');
        cy.contains('jwt').should('exist');
        cy.contains('https://urls.com').should('exist');
        cy.contains('https://trusted.com').should('exist');
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
        cy.contains('ui5-button', 'Edit').click();

        cy.contains(apiRuleName);

        // > CorsPolicy
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
        cy.get('[data-testid="spec.corsPolicy.allowHeaders.0"]')
            .find('input')
            .clear()
            .type("Allowed-Header")
        cy.get('[aria-label="expand CORS Allow Headers"]').should('exist').click();

        // CORS allow headers
        cy.get('[aria-label="expand CORS Expose Headers"]').should('be.visible').click();
        cy.get('[data-testid="spec.corsPolicy.exposeHeaders.0"]')
            .find('input')
            .clear()
            .type("Exposed-Header")
        cy.get('[aria-label="expand CORS Expose Headers"]').should('exist').click();

        // CORS allow credentials
        cy.get('ui5-switch[data-testid="spec.corsPolicy.allowCredentials"]')
            .find('[role="switch"]')
            .click();

        // CORS Max Age
        cy.get('[data-testid="spec.corsPolicy.maxAge"]')
            .find('input')
            .clear({force: true})
            .type('10s');

        cy.get('ui5-dialog')
            .contains('ui5-button', 'Update')
            .should('be.visible')
            .click();

        // Verify CORS policy
        cy.contains(apiRuleName).should("be.visible").click();
        cy.contains(apiRuleDefaultPath).should('exist');

        cy.contains('CORS Allow Methods').should('exist').parent().contains('GET').should('exist');
        cy.contains('CORS Expose Headers').should('exist').parent().contains('Exposed-Header').should('exist')
        cy.contains('CORS Allow Headers').should('exist').parent().contains('Allowed-Header').should('exist')
        cy.contains('CORS Allow Credentials').should('exist').parent().contains('true').should('exist')
        cy.contains('CORS Max Age').should('exist').parent().contains('10s').should('exist')
    });

});
