Cypress.Commands.add('createNamespace', (namespaceName) => {
    // Go to the details of namespace
    cy.getLeftNav()
        .contains('Namespaces')
        .click();

    cy.contains('ui5-button', 'Create Namespace').click();

    cy.get('ui5-input[aria-label="Namespace name"]')
        .find('input')
        .type(namespaceName, { force: true });

    cy.get('ui5-dialog')
        .contains('ui5-button', 'Create')
        .should('be.visible')
        .click();
});

Cypress.Commands.add('deleteNamespace', (namespaceName) => {
    cy.getLeftNav()
        .contains('Namespaces', { includeShadowDom: true })
        .click();

    cy.get('[role="search"] [aria-label="search-input"]').type(
        namespaceName,
        {
            force: true,
        },
    ); // use force to skip clicking (the table could re-render between the click and the typing)

    cy.get('tbody tr [aria-label="Delete"]').click({ force: true });

    cy.contains('button', 'Delete')
        .filter(':visible', { log: false })
        .click({ force: true });
});