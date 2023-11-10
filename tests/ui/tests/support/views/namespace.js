Cypress.Commands.add('createNamespace', (namespaceName) => {
    // Go to the details of namespace
    cy.getLeftNav()
        .contains('Namespaces', { includeShadowDom: true })
        .click();

    cy.contains('Create Namespace').click();

    cy.get('[role=dialog]')
        .find('input[ariaLabel="Namespace name"]:visible')
        .type(namespaceName);

    cy.get('[role=dialog]')
        .contains('button', 'Create')
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