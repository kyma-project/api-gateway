import { postApis } from "./httpClient";
import { loadFixture } from "./loadFile";

export type ApiRuleAccessStrategy = "oauth2_introspection" | "jwt" | "noop" | "allow" | "no_auth"

export type ApiRuleConfig = {
    name: string,
    namespace: string;
    service: string;
    host: string;
    handler: ApiRuleAccessStrategy;
    config?: JwtConfig | OAuth2IntroConfig | null;
    gateway?: string | null;
}

type JwtConfig = {
    jwks_urls: string[];
    trusted_issuers: string[];
}

type OAuth2IntroConfig = {
    required_scope: string[];
}

type ApiRule = {
    apiVersion: string;
    metadata: {
        name: string;
        namespace: string;
    }
    spec: {
        service: {
            name: string;
        }
        host: string;
        gateway: string;
        rules: {
            path: string;
            methods: string[];
            accessStrategies: {
                handler: ApiRuleAccessStrategy;
                config?: JwtConfig | OAuth2IntroConfig | null;
            }[];
        }[];
    }
}

Cypress.Commands.add('createApiRule', (cfg: ApiRuleConfig) => {
    // @ts-ignore Typing of cy.then is not good enough
    cy.wrap(loadFixture('apiRule.yaml')).then((a: ApiRule): void => {
        a.metadata.name = cfg.name;
        a.metadata.namespace = cfg.namespace;
        a.spec.service.name = cfg.service;
        a.spec.host = cfg.host;
        if (cfg.gateway != null) {
            a.spec.gateway = cfg.gateway;
        }
        a.spec.rules[0].accessStrategies = [
            {
                handler: cfg.handler,
                config: cfg.config
            }
        ]

        // We have to use cy.wrap, since the post command uses a cy.fixture internally
        cy.wrap(postApis(`${a.apiVersion}/namespaces/${cfg.namespace}/apirules`, a)).should("be.true");
    })
});