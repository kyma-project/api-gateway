import * as jsyaml from 'js-yaml';

export async function loadFixture(fileName: string, single = true) : Promise<Object> {
  const load = single ? jsyaml.load : jsyaml.loadAll;
  return new Promise(resolve => {
    // @ts-ignore Since type lib of js-yaml wasn't working correctly
    cy.fixture(fileName).then(fileContent => resolve(load(fileContent)));
  });
}