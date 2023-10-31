Cypress.Commands.add('navigateTo', (leftNav, resource) => {
  // To check and probably remove after cypress bump
  cy.wait(500);

  cy.getLeftNav()
    .contains(leftNav, { includeShadowDom: true })
    .should('be.visible');

  cy.getLeftNav()
    .contains(leftNav, { includeShadowDom: true })
    .click();

  cy.getLeftNav()
    .contains(resource, { includeShadowDom: true })
    .click();
});
