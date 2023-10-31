/// <reference types="cypress" />
import 'cypress-file-upload';

function openSearchWithSlashShortcut() {
  cy.get('body').type('/', { force: true });
}

const random = Math.floor(Math.random() * 9999) + 1000;
const FUNCTION_NAME = 'test-function';
const API_RULE_NAME = `test-api-rule-${random}`;
const API_RULE_PATH = '/test-path';
const API_RULE_DEFAULT_PATH = '/.*';

context('Test API Rules in the Function details view', () => {
  Cypress.skipAfterFail();

  before(() => {
    cy.loginAndSelectCluster();
    cy.goToNamespaceDetails();
  });

  it('Go to details of the simple Function', () => {
    cy.navigateTo('Workloads', 'Functions');

    cy.contains('a', FUNCTION_NAME)
      .filter(':visible', { log: false })
      .first()
      .click({ force: true });

    cy.get('[role="status"]').contains('span', /running/i, {
      timeout: 60 * 300,
    });
  });

  it('Create an API Rule for the Function', () => {
    cy.navigateTo('Discovery and Network', 'API Rules');

    cy.contains('Create API Rule').click();

    // Name
    cy.get('[ariaLabel="APIRule name"]:visible', { log: false }).type(
      API_RULE_NAME,
    );

    cy.get('[data-testid="spec.timeout"]:visible', { log: false })
      .clear()
      .type(1212);

    // Service
    cy.get('[aria-label="Choose Service"]:visible', { log: false })
      .first()
      .type(FUNCTION_NAME);

    cy.get('[aria-label="Choose Service"]:visible', { log: false })
      .first()
      .next()
      .find('[aria-label="Combobox input arrow"]:visible', { log: false })
      .click();

    cy.get('[data-testid="spec.service.port"]:visible', { log: false })
      .clear()
      .type(80);

    // Host
    cy.get('[aria-label="Combobox input"]:visible', { log: false })
      .first()
      .type('*');

    cy.get('span', { log: false })
      .contains(/^\*$/i)
      .first()
      .click();

    cy.get('[aria-label="Combobox input"]:visible', { log: false })
      .first()
      .type('{home}{rightArrow}{backspace}');

    cy.get('[aria-label="Combobox input"]:visible', { log: false })
      .first()
      .type(API_RULE_NAME);

    cy.get('[aria-label="Combobox input"]:visible', { log: false })
      .first()
      .next()
      .find('[aria-label="Combobox input arrow"]:visible', { log: false })
      .click();

    // Rules

    // > General

    cy.get('[data-testid="spec.rules.0.timeout"]:visible')
      .clear()
      .type(2323);

    // > Access Strategies

    cy.get('[data-testid="spec.rules.0.accessStrategies.0.handler"]:visible')
      .clear()
      .type('oauth2_introspection');

    cy.get('[data-testid="spec.rules.0.accessStrategies.0.handler"]:visible', {
      log: false,
    })
      .find('span')
      .find('[aria-label="Combobox input arrow"]:visible', { log: false })
      .click();

    cy.get('[aria-label="expand Required Scope"]:visible', {
      log: false,
    }).click();

    cy.get(
      '[data-testid="spec.rules.0.accessStrategies.0.config.required_scope.0"]:visible',
    )
      .clear()
      .type('read');

    // > Methods

    cy.get('[data-testid="spec.rules.0.methods.POST"]:visible').click();

    cy.get('[role=dialog]')
      .contains('button', 'Create')
      .click();
  });

  it('Check the API Rule details', () => {
    cy.contains(API_RULE_NAME).click();

    cy.contains(API_RULE_DEFAULT_PATH).should('exist');

    cy.contains('1212').should('exist');

    cy.contains('Rules #1', { timeout: 10000 }).click();

    cy.contains('2323').should('exist');

    cy.contains('oauth2_introspection').should('exist');

    cy.contains(API_RULE_PATH).should('not.exist');

    cy.contains('allow').should('not.exist');
    cy.contains('read').should('exist');
  });

  it('Edit the API Rule', () => {
    cy.contains('Edit').click();

    cy.contains(API_RULE_NAME);

    // Rules

    // > General

    cy.get('[aria-label="expand Rules"]:visible', { log: false })
      .contains('Add')
      .click();

    cy.get('[aria-label="expand Rule"]:visible', { log: false })
      .first()
      .click();

    cy.get('[data-testid="spec.rules.1.path"]:visible')
      .clear()
      .type(API_RULE_PATH);

    // > Access Strategies
    cy.get('[aria-label="expand Access Strategies"]:visible', { log: false })
      .first()
      .scrollIntoView();

    cy.get('[data-testid="spec.rules.1.accessStrategies.0.handler"]:visible')
      .find('input')
      .clear()
      .type('jwt');

    cy.get('[data-testid="spec.rules.1.accessStrategies.0.handler"]:visible', {
      log: false,
    })
      .find('span')
      .find('[aria-label="Combobox input arrow"]:visible', { log: false })
      .click();

    cy.get('[aria-label="expand JWKS URLs"]:visible', { log: false }).click();

    cy.get(
      '[data-testid="spec.rules.1.accessStrategies.0.config.jwks_urls.0"]:visible',
    )
      .clear()
      .type('https://urls.com');

    cy.get('[aria-label="expand Trusted Issuers"]:visible', {
      log: false,
    }).click();

    cy.get(
      '[data-testid="spec.rules.1.accessStrategies.0.config.trusted_issuers.0"]:visible',
    )
      .clear()
      .type('https://trusted.com');

    // > Methods
    cy.get('[data-testid="spec.rules.1.methods.GET"]:visible').click();

    cy.get('[data-testid="spec.rules.1.methods.POST"]:visible').click();

    cy.get('[role=dialog]')
      .contains('button', 'Update')
      .click();
  });

  it('Check the edited API Rule details', () => {
    cy.contains(API_RULE_NAME).click();

    cy.contains(API_RULE_DEFAULT_PATH).should('exist');

    cy.contains('Rules #1', { timeout: 10000 }).click();

    cy.contains('Rules #2', { timeout: 10000 }).click();

    cy.contains(API_RULE_PATH).should('exist');

    cy.contains('jwt').should('exist');
    cy.contains('https://urls.com').should('exist');
    cy.contains('https://trusted.com').should('exist');
  });

  it('Inspect list using slash shortcut', () => {
    cy.getLeftNav()
      .contains('API Rules', { includeShadowDom: true })
      .click();

    cy.contains('h3', 'API Rules').should('be.visible');
    cy.get('[aria-label="open-search"]').should('not.be.disabled');

    openSearchWithSlashShortcut();

    cy.get('[role="search"] [aria-label="search-input"]').type(API_RULE_NAME);

    cy.contains(API_RULE_NAME).should('be.visible');
  });

  it('Create OAuth2 Introspection rule', () => {
    cy.get('[class="fd-link"]')
      .contains(API_RULE_NAME)
      .click();

    cy.contains('Edit').click();

    cy.contains(API_RULE_NAME);

    // Rules
    cy.get('[aria-label="expand Rules"]:visible', { log: false })
      .contains('Add')
      .click();

    cy.get('[aria-label="expand Rule"]', { log: false })
      .first()
      .click();

    cy.get('[aria-label="expand Rule"]', { log: false })
      .eq(1)
      .click();

    cy.get('[data-testid="spec.rules.2.path"]:visible')
      .clear()
      .type(API_RULE_PATH);

    // > Access Strategies
    cy.get('[aria-label="expand Access Strategies"]:visible', { log: false })
      .first()
      .scrollIntoView();

    cy.get('[data-testid="spec.rules.2.accessStrategies.0.handler"]:visible')
      .find('input')
      .clear()
      .type('oauth2_introspection');

    cy.get('[data-testid="spec.rules.2.accessStrategies.0.handler"]:visible', {
      log: false,
    })
      .find('span')
      .find('[aria-label="Combobox input arrow"]:visible', { log: false })
      .click();

    cy.get('[aria-label="expand Introspection Request Headers"]:visible', {
      log: false,
    }).click();

    cy.get('[aria-label="expand Introspection Request Headers"]:visible', {
      log: false,
    })
      .parent()
      .within(_$div => {
        cy.get('[placeholder="Enter key"]:visible', { log: false })
          .first()
          .clear()
          .type('Authorization');

        cy.get('[placeholder="Enter value"]:visible', { log: false })
          .first()
          .clear()
          .type('Basic 12345');
      });

    cy.get(
      '[data-testid="spec.rules.2.accessStrategies.0.config.introspection_url"]:visible',
    )
      .clear()
      .type('https://example.com');

    cy.get('[aria-label="expand Token From"]:visible', {
      log: false,
    }).click();

    cy.get('[aria-label="expand Token From"]:visible', {
      log: false,
    })
      .parent()
      .within(_$div => {
        cy.contains('Enter key', { log: false })
          .first()
          .click();
      });

    cy.get('.fd-list__title')
      .contains('header')
      .click();

    cy.get('[aria-label="expand Token From"]:visible', {
      log: false,
    })
      .parent()
      .within(_$div => {
        cy.get('[placeholder="Enter value"]:visible', { log: false })
          .first()
          .clear()
          .type('FromHeader');
      });

    // > Methods
    cy.get('[data-testid="spec.rules.2.methods.GET"]:visible').click();

    cy.get('[data-testid="spec.rules.2.methods.POST"]:visible').click();

    cy.get('[role=dialog]')
      .contains('button', 'Update')
      .click();
  });

  it('Check OAuth2 Introspection strategy', () => {
    cy.contains(API_RULE_NAME).click();

    cy.contains(API_RULE_DEFAULT_PATH).should('exist');

    cy.contains('Rules #3', { timeout: 10000 }).click();

    cy.contains(API_RULE_PATH).should('exist');

    cy.contains('https://example.com').should('exist');
    cy.contains('Authorization=Basic 12345').should('exist');
    cy.contains('header=FromHeader').should('exist');
  });
});
