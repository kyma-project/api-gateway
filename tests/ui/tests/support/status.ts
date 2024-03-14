export type Status = "OK" | "ERROR" | "WARNING";

Cypress.Commands.add('hasStatusLabel', (status: Status) : void => {
    cy.get(`[aria-label="Status"]:visible`).contains(status);
});