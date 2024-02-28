Cypress.Commands.add('clickCreateButton', () => {
    cy.contains('ui5-button', 'Create').click();
});