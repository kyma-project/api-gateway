Cypress.Commands.add('inputClearAndType', (selector: string, newValue: string): void => {
    cy.get(selector)
        .find('input')
        .click()
        .clear({force: true})
        .type(newValue, {force: true});
});
