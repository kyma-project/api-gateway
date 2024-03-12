export type Status = "OK" | "ERROR" | "WARNING";

Cypress.Commands.add('hasStatusLabel', (status: Status) => {
    cy.get(`[aria-label="Status"]:visible`).contains(status);
});