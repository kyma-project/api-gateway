import {ApiRuleConfig} from "./apiRule";
import Chainable = Cypress.Chainable;

export interface Commands {
    createApiRule(cfg: ApiRuleConfig): void
    createService(name: string, namespace: string): void
    createNamespace(name: string): void
    deleteNamespace(name: string): void
}