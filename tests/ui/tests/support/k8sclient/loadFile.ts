import * as jsyaml from 'js-yaml';

export async function loadFixture(fileName: string, single = true) {
  const load = single ? jsyaml.load : jsyaml.loadAll;
  return new Promise(resolve => {
    cy.fixture(fileName).then(fileContent => resolve(load(fileContent)));
  });
}