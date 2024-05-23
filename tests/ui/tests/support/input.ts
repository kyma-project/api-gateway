Cypress.Commands.add('inputClearAndType', (selector: string, newValue: string): void => {
    cy.get(selector)
        .find('input')
        .scrollIntoView()
        .click({force: true})
        .clear({force: true})
        .type(newValue, {force: true});
});
