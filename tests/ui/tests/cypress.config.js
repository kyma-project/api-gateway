const { defineConfig } = require('cypress');

module.exports = defineConfig({
  defaultCommandTimeout: 60000,
  execTimeout: 60000,
  taskTimeout: 60000,
  pageLoadTimeout: 10000,
  requestTimeout: 10000,
  responseTimeout: 10000,
  fixturesFolder: 'fixtures',
  chromeWebSecurity: false,
  viewportWidth: 1500,
  viewportHeight: 1500,
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
      'tests/cluster/test-applications.spec.js',
      // 'tests/cluster/test-modules-wizard.spec.js',
      'tests/namespace/a-run-before.spec.js',
      'tests/namespace/run-prepare-functions.spec.js',
      'tests/namespace/test-certificates.spec.js',
      'tests/namespace/test-destination-rules.spec.js',
      'tests/namespace/test-issuers.spec.js',
      'tests/namespace/test-dns-providers.spec.js',
      'tests/namespace/test-virtual-services.spec.js',
      'tests/namespace/test-authorization-policies.spec.js',
      'tests/namespace/test-gateways.spec.js',
      'tests/namespace/test-hpa.spec.js',
      'tests/namespace/test-oauth2.spec.js',
      'tests/namespace/test-sidecars.spec.js',
      'tests/namespace/test-service-entries.spec.js',
      'tests/namespace/test-api-rules.spec.js',
      'tests/namespace/test-dns-entries.spec.js',
      'tests/namespace/test-services.spec.js',
      'tests/namespace/test-kyma.spec.js',
      // 'tests/namespace/test-module-templates.spec.js',
      'tests/namespace/z-run-after.spec.js',
    ],
    supportFile: 'support/index.js',
  },
});
