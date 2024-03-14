import config from './dashboard/config';

Cypress.Commands.add('loginAndSelectCluster', function () {

    sessionStorage.clear();
    cy.clearCookies();
    cy.clearLocalStorage();

    cy.visit(`${config.clusterAddress}/clusters`)
        .get('ui5-button:visible')
        .contains('Connect')
        .click();

    cy.get('input[type="file"]').attachFile('kubeconfig.yaml', {
        subjectType: 'drag-n-drop',
    });

    cy.contains('Next').click();

    cy.get(`[aria-label="next-step"]:visible`)
        .click({force: true});

    cy.get(`[aria-label="last-step"]:visible`)
        .contains('Connect')
        .click({force: true});

    cy.url().should('match', /overview$/);
    cy.contains('ui5-title', 'Cluster Details').should('be.visible');
});
