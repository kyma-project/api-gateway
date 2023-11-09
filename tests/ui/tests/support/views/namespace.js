Cypress.Commands.add('goToNamespaceDetails', () => {
    // Go to the details of namespace
    cy.getLeftNav()
        .contains('Namespaces', { includeShadowDom: true })
        .click();

    cy.get('[role=row]')
        .contains('a', Cypress.env('NAMESPACE_NAME'))
        .click();

    return cy.end();
});

Cypress.Commands.add('createNamespace', () => {
    // Go to the details of namespace
    cy.getLeftNav()
        .contains('Namespaces', { includeShadowDom: true })
        .click();

    cy.contains('Create Namespace').click();

    cy.get('[role=dialog]')
        .find('input[ariaLabel="Namespace name"]:visible')
        .type(Cypress.env('NAMESPACE_NAME'));

    cy.get('[role=dialog]')
        .contains('button', 'Create')
        .click();

    return cy.end();
});

Cypress.Commands.add('deleteNamespace', () => {
    cy.getLeftNav()
        .contains('Namespaces', { includeShadowDom: true })
        .click();

    cy.get('[role="search"] [aria-label="search-input"]').type(
        Cypress.env('NAMESPACE_NAME'),
        {
            force: true,
        },
    ); // use force to skip clicking (the table could re-render between the click and the typing)

    cy.get('tbody tr [aria-label="Delete"]').click({ force: true });

    cy.contains('button', 'Delete')
        .filter(':visible', { log: false })
        .click({ force: true });

    return cy.end();
});