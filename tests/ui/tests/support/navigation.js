Cypress.Commands.add('navigateTo', (leftNav, resource) => {
    // To check and probably remove after cypress bump
    cy.wait(500);

    cy.getLeftNav()
        .contains(leftNav)
        .should('be.visible');

    cy.getLeftNav()
        .contains(leftNav)
        .click();

    cy.getLeftNav()
        .contains(resource)
        .click();
});

Cypress.Commands.add('getLeftNav', () => {
    return cy.get('aside', { timeout: 10000 });
});