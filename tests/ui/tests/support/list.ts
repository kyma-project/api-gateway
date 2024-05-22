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

Cypress.Commands.add('hasTableRowWithLink', (hrefValue: string) : void => {
    cy.get('ui5-table-row')
        .find('ui5-link')
        .should('have.attr', 'href', hrefValue)
});

Cypress.Commands.add('hasTableRowNumberWithLink', (row: number, hrefValue: string) : void => {
    cy.get(`ui5-table-row[slot="default-${row}"]`)
        .find('ui5-link')
        .should('have.attr', 'href', hrefValue)
});
