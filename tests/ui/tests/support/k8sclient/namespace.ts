import * as k8s from '@kubernetes/client-node';

import {deleteResourceApiEndpoint, post, postApiEndpoint} from "./httpClient";
import {loadFixture} from "./loadFile";
import config from "../../config";

Cypress.Commands.add('createNamespace', function (name: string) {
    cy.wrap(loadFixture('namespace.yaml')).then((ns: k8s.V1Namespace) => {
        ns.metadata.name = name
        // We have to use cy.wrap, since the post command uses a cy.fixture internally
        cy.wrap(postApiEndpoint('v1/namespaces', ns)).should("be.true");
    })
})

Cypress.Commands.add('deleteNamespace', function (name: string) {
    // We have to use cy.wrap, since the post command uses a cy.fixture internally
    cy.wrap(deleteResourceApiEndpoint(`v1/namespaces/${name}`)).should("be.true");
})
