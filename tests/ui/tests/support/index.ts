import './exceptions';
import './login';
import './navigation';
import './k8sclient';
import './random';
import './list';
import './buttons';
import './combobox';
import {K8sClient} from "./k8sclient";

declare global {
    namespace Cypress {
        interface Chainable extends K8sClient {
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
        }
    }
}