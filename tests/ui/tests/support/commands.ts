import {K8sClientCommands} from "./k8sclient";
import {Status} from "./status";
import {ApiRuleCommands} from "./apirule";
import {ButtonCommands} from "./buttons";
import {NavigationCommands} from "./navigation";

declare global {
    namespace Cypress {
        interface Chainable extends K8sClientCommands, ApiRuleCommands, ButtonCommands, NavigationCommands {
            loginAndSelectCluster(): void
            clickGenericListLink(resourceName: string): void
            chooseComboboxOption(selector: string, optionText: string): void
            filterWithNoValue(): Chainable<JQuery>
            inputClearAndType(selector: string, newValue: string): void
            hasStatusLabel(status: Status): void
            hasTableRowWithLink(hrefValue: string): void
        }
    }
}