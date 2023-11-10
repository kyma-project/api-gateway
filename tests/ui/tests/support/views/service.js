Cypress.Commands.add('createService', serviceName => {
  cy.navigateTo('Discovery and Network', 'Services');

  cy.contains('Create Service').click();

  cy.get('[ariaLabel="Service name"]:visible', { log: false }).type(
      serviceName,
  );

  cy.get('[role="dialog"]')
      .contains('button', 'Add')
      .click();

  cy.get('[ariaLabel="Service name"]:visible', { log: false })
      // Because the port name field has the same ariaLabel as the Service name field, we have
      // to use the second element to fill in the port name.
      .eq(1)
      .type(serviceName);

  cy.get('[role="dialog"]')
    .contains('button', 'Create')
    .click();
});
