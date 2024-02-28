import * as jsyaml from 'js-yaml';

export function loadFile(FILE_NAME, single = true) {
  const load = single ? jsyaml.load : jsyaml.loadAll;
  return new Promise(resolve => {
    cy.fixture(FILE_NAME).then(fileContent => resolve(load(fileContent)));
  });
}

