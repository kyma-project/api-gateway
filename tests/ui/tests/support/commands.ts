import {K8sClientCommands} from "./k8sclient";
import {Status} from "./status";
import {ApiRuleCommands} from "./apirule";
import {ButtonCommands} from "./buttons";
import {NavigationCommands} from "./navigation";

declare global {
    namespace Cypress {
        interface Chainable extends K8sClientCommands, ApiRuleCommands, ButtonCommands, NavigationCommands {
            loginAndSelectCluster(): Chainable<JQuery>
            clickGenericListLink(resourceName: string): Chainable<JQuery>
            chooseComboboxOption(selector: string, optionText: string): Chainable<JQuery>
            filterWithNoValue(): Chainable<JQuery>
            inputClearAndType(selector: string, newValue: string): Chainable<JQuery>
            hasStatusLabel(status: Status): Chainable<JQuery>
        }
    }
}