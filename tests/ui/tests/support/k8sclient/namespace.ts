import * as k8s from '@kubernetes/client-node';

import {deleteResource, postApi} from "./httpClient";
import {loadFixture} from "./loadFile";

Cypress.Commands.add('createNamespace', (name: string) => {
    // @ts-ignore Typing of cy.then is not good enough
    cy.wrap(loadFixture('namespace.yaml')).then((ns: k8s.V1Namespace): void => {
        ns.metadata!.name = name
        // We have to use cy.wrap, since the post command uses a cy.fixture internally
        cy.wrap(postApi('v1/namespaces', ns)).should("be.true");
    })
})

Cypress.Commands.add('deleteNamespace', (name: string) => {
    // We have to use cy.wrap, since the post command uses a cy.fixture internally
    cy.wrap(deleteResource(`v1/namespaces/${name}`)).should("be.true");
})
