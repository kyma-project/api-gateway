Cypress.Commands.add('inputClearAndType', (selector: string, newValue: string): void => {
    cy.get(selector)
        .find('input')
        .as('inpt')
        .scrollIntoView()
        .click({ force: true })
    cy.get('@inpt')
        .clear({ force: true })
        .type(newValue, { force: true });
});
