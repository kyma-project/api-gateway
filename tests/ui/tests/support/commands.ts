import {Commands} from "./k8sclient";
import {Status} from "./status";

declare global {
    namespace Cypress {
        interface Chainable extends Commands {
            loginAndSelectCluster(): Chainable<JQuery>
            getLeftNav(): Chainable<JQuery>
            clickCreateButton(): Chainable<JQuery>
            clickGenericListLink(resourceName: string): Chainable<JQuery>
            deleteFromGenericList(resourceName: string): Chainable<JQuery>
            navigateTo(leftNav: string, resource: string): Chainable<JQuery>
            navigateToNamespace(name: string): Chainable<JQuery>
            navigateToApiRule(name: string, namespace: string): Chainable<JQuery>
            navigateToApiRuleList(name: string): Chainable<JQuery>
            chooseComboboxOption(selector: string, optionText: string): Chainable<JQuery>
            filterWithNoValue(): Chainable<JQuery>
            hasStatusLabel(status: Status): Chainable<JQuery>
        }
    }
}