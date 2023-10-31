Cypress.Commands.add('checkExtension', (resource, create = true) => {
  cy.getLeftNav()
    .contains(resource, { includeShadowDom: true })
    .click();

  cy.get('[aria-label="title"]')
    .contains(resource)
    .should('be.visible');

  if (create) {
    cy.get('[type=button]')
      .contains('Create')
      .should('be.visible');
  }
});
