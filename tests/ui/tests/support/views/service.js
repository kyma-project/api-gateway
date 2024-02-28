Cypress.Commands.add('createService', (serviceName) => {
  cy.navigateTo('Discovery and Network', 'Services');

  cy.clickCreateButton();

  cy.get('ui5-input[aria-label="Service name"]')
      .find('input')
      .type(serviceName, { force: true });

  cy.get('ui5-dialog')
      .contains('ui5-button', 'Add')
      .should('be.visible')
      .click();

  cy.get('ui5-input[aria-label="Service name"]')
      // Because the port name field has the same ariaLabel as the Service name field, we have
      // to use the second element to fill in the port name.
      .eq(1)
      .find('input')
      .type(serviceName, { force: true });

  cy.get('ui5-dialog')
      .contains('ui5-button', 'Create')
      .should('be.visible')
      .click();
});
