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