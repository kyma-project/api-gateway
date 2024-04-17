const defaultKymaDashboardAddress = "http://localhost:3001";
const dashboardUrl = (Cypress.env("DOMAIN") as string) || defaultKymaDashboardAddress

export default {
  clusterAddress: dashboardUrl,
  backendApiUrl: `${dashboardUrl}/backend/api/`,
  backendApisUrl: `${dashboardUrl}/backend/apis/`,
};
