export function chooseComboboxOption(selector, optionText) {
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

    return cy.end();
}