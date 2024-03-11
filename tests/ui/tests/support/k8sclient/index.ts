import "./namespace"
import "./service"
import "./auth"
import "./loadFile"
import "./apiRule";
import {ApiRuleConfig} from "./apiRule";
import Chainable = Cypress.Chainable;

export interface K8sClient {
    createApiRule(cfg: ApiRuleConfig): Chainable<JQuery>
    createService(name: string, namespace: string): Chainable<JQuery>
    createNamespace(name: string): Chainable<JQuery>
    deleteNamespace(name: string): Chainable<JQuery>
}