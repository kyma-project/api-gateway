import Chainable = Cypress.Chainable;
import {ApiRuleAccessStrategy} from "./k8sclient";

export interface ApiRuleCommands {
    apiRuleTypeName(name: string): Chainable<JQuery>
    apiRuleSelectService(name: string): Chainable<JQuery>
    apiRuleTypeServicePort(port: string): Chainable<JQuery>
    apiRuleTypeHost(host: string): Chainable<JQuery>
    apiRuleSelectAccessStrategy(strategy: ApiRuleAccessStrategy, ruleNumber?: number): Chainable<JQuery>
    apiRuleTypeRulePath(path: string, ruleNumber?: number): Chainable<JQuery>
    apiRuleSelectMethod(method: ApiRuleMethods, ruleNumber?: number): Chainable<JQuery>
    apiRuleTypeJwksUrl(url: string, ruleNumber?: number): Chainable<JQuery>
    apiRuleTypeTrustedIssuer(url: string, ruleNumber?: number): Chainable<JQuery>
}

Cypress.Commands.add('apiRuleTypeName', (name: string) => {
    cy.inputClearAndType('ui5-input[aria-label="APIRule name"]', name);
});

Cypress.Commands.add('apiRuleSelectService', (name: string) => {
    cy.chooseComboboxOption('[data-testid="spec.service.name"]', name);
});

Cypress.Commands.add('apiRuleTypeServicePort', (port: string) => {
    cy.inputClearAndType('[data-testid="spec.service.port"]:visible', port);
});

Cypress.Commands.add('apiRuleTypeHost', (host: string) => {
    cy.inputClearAndType('[data-testid="spec.host"]:visible', host);
});

Cypress.Commands.add('apiRuleTypeRulePath', (path: string, ruleNumber: number = 0) => {
    cy.inputClearAndType(`[data-testid="spec.rules.${ruleNumber}.path"]:visible`, path);
});

Cypress.Commands.add('apiRuleSelectAccessStrategy', (strategy: ApiRuleAccessStrategy, ruleNumber: number = 0) => {
    cy.inputClearAndType(`ui5-combobox[data-testid="spec.rules.${ruleNumber}.accessStrategies.0.handler"]:visible`, strategy);

    cy.get('ui5-li:visible')
        .contains(strategy)
        .find('li')
        .click({force: true});
});

type ApiRuleMethods = "GET" | "POST" | "PUT" | "DELETE" | "PATCH" | "HEAD" | "OPTIONS" | "CONNECT" | "TRACE";
Cypress.Commands.add('apiRuleSelectMethod', (method: ApiRuleMethods, ruleNumber: number = 0) => {
    cy.get(`[data-testid="spec.rules.${ruleNumber}.methods.${method}"]`)
    cy.get(`ui5-checkbox[text="${method}"]:visible`).click();
});

Cypress.Commands.add('apiRuleTypeJwksUrl', (url: string, ruleNumber: number = 0) => {
    cy.get('[aria-label="expand JWKS URLs"]:visible', {log: false}).click();
    cy.inputClearAndType(`[data-testid="spec.rules.${ruleNumber}.accessStrategies.0.config.jwks_urls.0"]:visible`, url);
});

Cypress.Commands.add('apiRuleTypeTrustedIssuer', (url: string, ruleNumber: number = 0) => {
    cy.get('[aria-label="expand Trusted Issuers"]:visible', {log: false}).click();
    cy.inputClearAndType(`[data-testid="spec.rules.${ruleNumber}.accessStrategies.0.config.trusted_issuers.0"]:visible`, url);
});