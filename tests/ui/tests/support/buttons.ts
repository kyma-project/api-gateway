import Chainable = Cypress.Chainable;

export interface ButtonCommands {
    clickCreateButton(): Chainable<JQuery>
    clickEditButton(): Chainable<JQuery>
    clickDialogCreateButton(): Chainable<JQuery>
    clickDialogUpdateButton(): Chainable<JQuery>
}

Cypress.Commands.add('clickCreateButton', (): void => {
    cy.contains('ui5-button', 'Create')
        .should('be.visible')
        .click();
});

Cypress.Commands.add('clickDialogCreateButton', (): void => {
    cy.get('ui5-dialog')
        .contains('ui5-button', 'Create')
        .should('be.visible')
        .click();
});

Cypress.Commands.add('clickEditButton', (): void => {
    cy.contains('ui5-button', 'Edit')
        .should('be.visible')
        .click();
});

Cypress.Commands.add('clickDialogUpdateButton', (): void => {
    cy.get('ui5-dialog')
        .contains('ui5-button', 'Update')
        .should('be.visible')
        .click();
});