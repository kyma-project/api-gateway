Cypress.Commands.add('inputClearAndType', (selector: string, newValue: string): void => {
    cy.get(selector)
        .find('input')
        .scrollIntoView()
        .click({ force: true })
        .clear({ force: true })
        .type(newValue, { force: true });
});

Cypress.Commands.add('inputPairClearAndType', (selector: string, newValueFirst: string, newValueLast: string): void => {
    cy.get(selector)
        .find('input')
        .first()
        .scrollIntoView()
        .click({ force: true })
        .clear({ force: true })
        .type(newValueFirst, { force: true });

    cy.get(selector)
        .find('input')
        .last()
        .scrollIntoView()
        .click({ force: true })
        .clear({ force: true })
        .type(newValueLast, { force: true });
});
