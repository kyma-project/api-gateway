import Chainable = Cypress.Chainable;

Cypress.Commands.add('chooseComboboxOption', (selector: string, optionText: string) : void => {
    cy.get(`ui5-combobox${selector}:visible`)
        .find('input:visible')
        .filterWithNoValue()
        .click({ force: true })
        .type(optionText, { force: true });
    cy.wait(200);
    cy.contains(optionText)
        .find('li')
        .click({ force: true });
});

Cypress.Commands.add('filterWithNoValue', { prevSubject: true }, (subjects: Chainable<JQuery<HTMLInputElement>>): Chainable<JQuery> => {
    return subjects.filter((_, e) => !(e as HTMLInputElement).value)
});
