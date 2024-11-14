import Chainable = Cypress.Chainable;
import {ApiRuleAccessStrategy} from "./k8sclient";

export interface ApiRuleCommands {
    apiRuleTypeName(name: string): void
    apiRuleSelectService(name: string): void
    apiRuleTypeServicePort(port: string): void
    apiRuleTypeHost(host: string): void
    apiRuleSelectAccessStrategy(strategy: ApiRuleAccessStrategy, ruleNumber?: number): void
    apiRuleTypeRulePath(path: string, ruleNumber?: number): void
    apiRuleSelectMethod(method: ApiRuleMethods, ruleNumber?: number): void
    apiRuleTypeJwksUrl(url: string, ruleNumber?: number): void
    apiRuleTypeTrustedIssuer(url: string, ruleNumber?: number): void
    apiRuleMetadataContainsHostUrl(url: string): void
}

Cypress.Commands.add('apiRuleTypeName', (name: string): void => {
    cy.inputClearAndType('ui5-input[aria-label="APIRule name"]', name);
});

Cypress.Commands.add('apiRuleSelectService', (name: string): void => {
    cy.chooseComboboxOption('[data-testid="spec.service.name"]', name);
});

Cypress.Commands.add('apiRuleTypeServicePort', (port: string): void => {
    cy.inputClearAndType('[data-testid="spec.service.port"]:visible', port);
});

Cypress.Commands.add('apiRuleTypeHost', (host: string): void => {
    cy.inputClearAndType('[data-testid="spec.host"]:visible', host);
});

Cypress.Commands.add('apiRuleTypeRulePath', (path: string, ruleNumber: number = 0): void => {
    cy.inputClearAndType(`[data-testid="spec.rules.${ruleNumber}.path"]:visible`, path);
});

Cypress.Commands.add('apiRuleSelectAccessStrategy', (strategy: ApiRuleAccessStrategy, ruleNumber: number = 0): void => {
    cy.inputClearAndType(`ui5-combobox[data-testid="spec.rules.${ruleNumber}.accessStrategies.0.handler"]:visible`, strategy);

    cy.contains(strategy)
        .find('li')
        .click({force: true});
});

type ApiRuleMethods = "GET" | "POST" | "PUT" | "DELETE" | "PATCH" | "HEAD" | "OPTIONS" | "CONNECT" | "TRACE";
Cypress.Commands.add('apiRuleSelectMethod', (method: ApiRuleMethods, ruleNumber: number = 0): void => {
    cy.get(`[data-testid="spec.rules.${ruleNumber}.methods.${method}"]`)
    cy.get(`ui5-checkbox[text="${method}"]:visible`).click();
});

Cypress.Commands.add('apiRuleTypeJwksUrl', (url: string, ruleNumber: number = 0): void => {
    cy.get('[aria-label="expand JWKS URLs"]:visible', {log: false}).click();
    cy.inputClearAndType(`[data-testid="spec.rules.${ruleNumber}.accessStrategies.0.config.jwks_urls.0"]:visible`, url);
});

Cypress.Commands.add('apiRuleTypeTrustedIssuer', (url: string, ruleNumber: number = 0): void => {
    cy.get('[aria-label="expand Trusted Issuers"]:visible', {log: false}).click();
    cy.inputClearAndType(`[data-testid="spec.rules.${ruleNumber}.accessStrategies.0.config.trusted_issuers.0"]:visible`, url);
});

Cypress.Commands.add('apiRuleMetadataContainsHostUrl', (url: string): void => {
    cy.get('ui5-card')
        .find('ui5-link')
        .should('have.attr', 'href', url)
});