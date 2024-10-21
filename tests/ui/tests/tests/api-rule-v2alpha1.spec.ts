import 'cypress-file-upload';
import {generateNamespaceName, generateRandomName} from "../support";

context("API Rule v2alpha1", () => {

    let apiRuleName = "";
    let namespaceName = "";
    let serviceName = "";

    beforeEach(() => {
        apiRuleName = generateRandomName("test-api-rule");
        namespaceName = generateNamespaceName();
        serviceName = generateRandomName("test-service");
        cy.loginAndSelectCluster();
        cy.createNamespace(namespaceName);
        cy.createService(serviceName, namespaceName);
    });

    afterEach(() => {
        cy.deleteNamespace(namespaceName);
    });

    it("should not display v1beta1 APIRule in v2alpha1 list", () => {
        cy.createApiRule({
            name: apiRuleName,
            namespace: namespaceName,
            service: serviceName,
            host: apiRuleName,
            handler: "no_auth",
        });

        cy.navigateToApiRuleList(namespaceName);
        cy.contains(apiRuleName).should('exist');

        cy.navigateToApiRuleV2alpha1List(namespaceName);
        cy.contains(apiRuleName).should('not.exist');
    });

    it('should not display v2alpha1 APIRule in v1beta1 list', () => {

        cy.createApiRuleV2alpha1({
            name: apiRuleName,
            namespace: namespaceName,
            service: serviceName,
            host: apiRuleName,
        });

        cy.navigateToApiRuleV2alpha1List(namespaceName);
        cy.contains(apiRuleName).should('exist');

        cy.navigateToApiRuleList(namespaceName);
        cy.contains(apiRuleName).should('not.exist');
    });

});
