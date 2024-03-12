Cypress.Commands.add('chooseComboboxOption', (selector: string, optionText: string) : void => {
    cy.get(`ui5-combobox${selector}:visible`)
        .find('input:visible')
        .filterWithNoValue()
        .click({ force: true })
        .type(optionText, { force: true });
    cy.wait(200);
    cy.get('ui5-li:visible', { timeout: 10000 })
        .contains(optionText)
        .find('li')
        .click({ force: true });
});

Cypress.Commands.add('filterWithNoValue', { prevSubject: true }, $elements =>
    $elements.filter((_, e) => !e.value),
);