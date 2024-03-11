import * as k8s from '@kubernetes/client-node';

import {deleteResource, post} from "./httpClient";
import {loadFixture} from "./loadFile";

Cypress.Commands.add('createNamespace', function (name: string) {
    cy.wrap(loadFixture('namespace.yaml')).then((ns: k8s.V1Namespace) => {
        ns.metadata.name = name
        // We have to use cy.wrap, since the post command uses a cy.fixture internally
        cy.wrap(post('v1/namespaces', ns)).should("be.true");
    })
})

Cypress.Commands.add('deleteNamespace', function (name: string) {
    // We have to use cy.wrap, since the post command uses a cy.fixture internally
    cy.wrap(deleteResource(`v1/namespaces/${name}`)).should("be.true");
})
