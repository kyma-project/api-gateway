Cypress.Commands.add('filterWithNoValue', { prevSubject: true }, $elements =>
    $elements.filter((_, e) => !e.value),
);