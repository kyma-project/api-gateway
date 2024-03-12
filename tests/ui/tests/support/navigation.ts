import {getK8sCurrentContext} from "./k8sclient";
import config from "./dashboard/config";
import Chainable = Cypress.Chainable;

export interface NavigationCommands {
    getLeftNav(): Chainable<JQuery>
    navigateTo(leftNav: string, resource: string): Chainable<JQuery>
    navigateToApiRule(name: string, namespace: string): Chainable<JQuery>
    navigateToApiRuleList(name: string): Chainable<JQuery>
}

Cypress.Commands.add('navigateTo', (leftNav: string, resource: string) : void => {
    // To check and probably remove after cypress bump
    cy.wait(500);

    cy.getLeftNav()
        .contains(leftNav)
        .should('be.visible');

    cy.getLeftNav()
        .contains(leftNav)
        .click();

    cy.getLeftNav()
        .contains(resource)
        .click();
});

Cypress.Commands.add('getLeftNav', () : void => {
    cy.get('aside', { timeout: 10000 });
});

Cypress.Commands.add('navigateToApiRule', (name: string, namespace: string) : void => {
    cy.wrap(getK8sCurrentContext()).then((context) => {
        cy.visit(`${config.clusterAddress}/cluster/${context}/namespaces/${namespace}/apirules/${name}`)
        cy.wait(2000);
    });
});

Cypress.Commands.add('navigateToApiRuleList', (namespace: string) : void => {
    cy.wrap(getK8sCurrentContext()).then((context) => {
        cy.visit(`${config.clusterAddress}/cluster/${context}/namespaces/${namespace}/apirules`)
        cy.wait(2000);
    });
});