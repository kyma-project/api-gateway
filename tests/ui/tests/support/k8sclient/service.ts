import {loadFixture} from "./loadFile";
import * as k8s from "@kubernetes/client-node";
import {postApi} from "./httpClient";

Cypress.Commands.add('createService', (name: string, namespace: string) => {
  cy.wrap(loadFixture('service.yaml')).then((s: k8s.V1Service) => {
    s.metadata.name = name
    // We have to use cy.wrap, since the post command uses a cy.fixture internally
    cy.wrap(postApi(`v1/namespaces/${namespace}/services`, s)).should("be.true");
  })
});


