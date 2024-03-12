import * as jsyaml from 'js-yaml';
import {readFileSync} from "fs";

export async function loadFixture(fileName: string, single = true) : Promise<Object> {
  const load = single ? jsyaml.load : jsyaml.loadAll;
  return new Promise(resolve => {
    cy.fixture(fileName).then(fileContent => resolve(load(fileContent)));
  });
}