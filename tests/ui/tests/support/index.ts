import './exceptions';
import './login';
import './navigation';
import './loadFile';
import './views';
import './random';
import './list';
import './buttons';
import './combobox';

declare global {
    namespace Cypress {
        interface Chainable {
            createNamespace(name: string): Chainable<JQuery>
            deleteNamespace(name: string): Chainable<JQuery>
            loginAndSelectCluster(): Chainable<JQuery>
            createService(name: string): Chainable<JQuery>
            getLeftNav(): Chainable<JQuery>
            clickCreateButton(): Chainable<JQuery>
            clickGenericListLink(resourceName: string): Chainable<JQuery>
            deleteFromGenericList(resourceName: string): Chainable<JQuery>
            navigateTo(leftNav: string, resource: string): Chainable<JQuery>
            loadFile(fileName: string): Chainable<JQuery>
            chooseComboboxOption(selector: string, optionText: string): Chainable<JQuery>
            filterWithNoValue(): Chainable<JQuery>
        }
    }
}