const { defineConfig } = require('cypress');

module.exports = defineConfig({
  includeShadowDom: true,
  defaultCommandTimeout: 20000,
  execTimeout: 60000,
  taskTimeout: 60000,
  pageLoadTimeout: 10000,
  requestTimeout: 10000,
  responseTimeout: 10000,
  fixturesFolder: 'fixtures',
  chromeWebSecurity: false,
  viewportWidth: 1500,
  viewportHeight: 1500,
  video: true,
  videoCompression: false,
  screenshotsFolder: process?.env?.ARTIFACTS
    ? `${process.env?.ARTIFACTS}/screenshots`
    : 'cypress/screenshots',
  videosFolder: process?.env?.ARTIFACTS
    ? `${process.env?.ARTIFACTS}/videos`
    : 'cypress/videos',
  experimentalInteractiveRunEvents: true,
  numTestsKeptInMemory: 0,
  e2e: {
    testIsolation: false,
    experimentalRunAllSpecs: true,
    setupNodeEvents(on, config) {
      return require('./plugins')(on, config);
    },
    specPattern: [
      'tests/**/*.spec.js',
    ],
    supportFile: 'support/index.js',
  },
});
