const defaultKymaDashboardAddress = "http://localhost:3001";


export default {
  clusterAddress: (Cypress.env("DOMAIN") as string) || defaultKymaDashboardAddress
};
