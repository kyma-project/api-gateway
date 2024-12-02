import {getK8sCurrentContext} from "./k8sclient";
import config from "./dashboard/config";
import Chainable = Cypress.Chainable;

export interface NavigationCommands {
    navigateToApiRule(name: string, namespace: string): void
    navigateToApiRuleList(name: string): void
    navigateToApiRuleV2alpha1List(name: string): void
}

Cypress.Commands.add('navigateToApiRule', (name: string, namespace: string) : void => {
    cy.wrap(getK8sCurrentContext()).then((context) => {
        cy.visit(`${config.clusterAddress}/cluster/${context}/namespaces/${namespace}/apirules/${name}`)
    });
    // Waiting to avoid dashboard rendering timing issues
    cy.wait(5000);
});

Cypress.Commands.add('navigateToApiRuleList', (namespace: string) : void => {
    cy.wrap(getK8sCurrentContext()).then((context) => {
        cy.visit(`${config.clusterAddress}/cluster/${context}/namespaces/${namespace}/apirules`)
    });
    // Waiting to avoid dashboard rendering timing issues
    cy.wait(5000);
});

Cypress.Commands.add('navigateToApiRuleV2alpha1List', (namespace: string) : void => {
    cy.wrap(getK8sCurrentContext()).then((context) => {
        cy.visit(`${config.clusterAddress}/cluster/${context}/namespaces/${namespace}/apirules-v2alpha1`)
    });
    // Waiting to avoid dashboard rendering timing issues
    cy.wait(5000);
    cy.contains('API Rules v2alpha1').should('exist');
});