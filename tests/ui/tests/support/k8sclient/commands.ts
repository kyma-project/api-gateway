import {ApiRuleConfig, ApiRuleV2alpha1Config} from "./apiRule";
import Chainable = Cypress.Chainable;

export interface Commands {
    createApiRule(cfg: ApiRuleConfig): void
    createApiRuleV2alpha1(cfg: ApiRuleV2alpha1Config): void
    createService(name: string, namespace: string): void
    createNamespace(name: string): void
    deleteNamespace(name: string): void
}