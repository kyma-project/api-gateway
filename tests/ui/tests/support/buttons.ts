

export interface ButtonCommands {
    clickCreateButton(): void
    clickEditTab(): void
    clickViewTab(): void
    clickSaveButton(): void
}

Cypress.Commands.add('clickCreateButton', (): void => {
    cy.contains('ui5-button', 'Create')
        .should('be.visible')
        .click();
});

Cypress.Commands.add('clickEditTab', (): void => {
    cy.get('ui5-tabcontainer')
        .contains('span', 'Edit')
        .should('be.visible')
        .click();
});

Cypress.Commands.add('clickViewTab', (): void => {
    cy.get('ui5-tabcontainer')
        .contains('span', 'View')
        .should('be.visible')
        .click();
});

Cypress.Commands.add('clickSaveButton', (): void => {
    cy.get('.edit-form')
        .contains('ui5-button:visible', 'Save')
        .click();
});