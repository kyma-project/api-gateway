import {getK8sCurrentContext} from "./k8sclient/auth";
import config from "../config";

Cypress.Commands.add('navigateTo', (leftNav, resource) => {
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

Cypress.Commands.add('getLeftNav', () => {
    return cy.get('aside', { timeout: 10000 });
});

Cypress.Commands.add('navigateToNamespace', (name: string) => {
    cy.wrap(getK8sCurrentContext()).then((context) => {
        cy.visit(`${config.clusterAddress}/cluster/${context}/namespaces/${name}`)
        cy.wait(2000);
    });
});

Cypress.Commands.add('navigateToApiRule', (name: string, namespace: string) => {
    cy.wrap(getK8sCurrentContext()).then((context) => {
        cy.visit(`${config.clusterAddress}/cluster/${context}/namespaces/${namespace}/apirules/${name}`)
        cy.wait(2000);
    });
});

Cypress.Commands.add('navigateToApiRuleList', (namespace: string) => {
    cy.wrap(getK8sCurrentContext()).then((context) => {
        cy.visit(`${config.clusterAddress}/cluster/${context}/namespaces/${namespace}/apirules`)
        cy.wait(2000);
    });
});