Cypress.Commands.add('clickGenericListLink', (resourceName: string) : void => {
    cy.get('ui5-table-row')
        .find('ui5-table-cell')
        .contains('span', resourceName)
        .should('be.visible')

    cy.get('ui5-table-row')
        .find('ui5-table-cell')
        .contains('span', resourceName)
        .click();
});

Cypress.Commands.add('hasTableRowNumberContaining', (row: number, value: string) : void => {
    cy.get(`ui5-table-row[slot="default-${row}"]`)
        .contains(value)
        .should('be.visible')
});
