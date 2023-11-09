Cypress.Commands.add('navigateToFunctionCreate', functionName => {
  cy.getLeftNav()
    .contains('Functions', { includeShadowDom: true })
    .click();

  cy.contains('Create Function').click();

  cy.get('[role="document"]').as('form');

  cy.get('@form')
    .find('.fd-select__text-content:visible')
    .contains('Choose preset')
    .click();

  cy.get('[role="list"]')
    .contains('Node.js Function')
    .click();

  cy.contains('Advanced').click();

  cy.get('.advanced-form')
    .find('[ariaLabel="Function name"]')
    .type(functionName);
});

Cypress.Commands.add('createSimpleFunction', functionName => {
  cy.getLeftNav()
    .contains('Workloads', { includeShadowDom: true })
    .click();

  cy.navigateToFunctionCreate(functionName);

  cy.get('[role="dialog"]')
    .contains('button', 'Create')
    .click();
});
