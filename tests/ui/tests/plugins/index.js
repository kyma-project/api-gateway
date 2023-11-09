const fs = require('fs');

module.exports = (on, config) => {
  let namespaceName = process.env.NAMESPACE_NAME || null;
  // generate random namespace name if it wasn't provided as env
  const random = Math.floor(Math.random() * 9999) + 1000;
  const randomName = `a-busola-test-${random}`;
  if (!namespaceName) {
    namespaceName = randomName;
  }
  const dynamicSharedStore = {
    cancelTests: false,
  };

  config.env.NAMESPACE_NAME = namespaceName;
  config.env.STORAGE_CLASS_NAME = randomName;
  config.env.APP_NAME = randomName;

  on('after:spec', (spec, results) => {
    if (results && results.video) {
      // Do we have failures for any retry attempts?
      const failures = results.tests.some((test) =>
          test.attempts.some((attempt) => attempt.state === 'failed')
      )
      if (!failures) {
        // delete the video if the spec passed and no tests retried
        fs.unlinkSync(results.video)
      }
    }
  });

  on('task', {

    // invoke setter cy.task('dynamicSharedStore', { name: 'cancelTests', value: true })
    // invoke getter cy.task('dynamicSharedStore', { name: 'cancelTests' })
    dynamicSharedStore(property) {
      if (property.value !== undefined) {
        return (dynamicSharedStore[property.name] = property.value);
      } else {
        return dynamicSharedStore[property.name];
      }
    },
  });
  return config;
};
