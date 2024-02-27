Cypress.Commands.add('createNamespace', (namespaceName) => {
    // Go to the details of namespace
    cy.getLeftNav()
        .contains('Namespaces')
        .click();

    cy.contains('ui5-button', 'Create').click();

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

    cy.deleteFromGenericList(namespaceName);

    cy.contains(`delete Namespace ${namespaceName}`);
    cy.get(`[header-text="Delete Namespace"]`)
        .find('[data-testid="delete-confirmation"]')
        .click();
});
