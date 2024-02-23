Cypress.Commands.add('clickGenericListLink', resourceName => {
    cy.get('ui5-table-row')
        .find('ui5-table-cell')
        .find('ui5-link')
        .contains(resourceName)
        .click({ force: true });
});