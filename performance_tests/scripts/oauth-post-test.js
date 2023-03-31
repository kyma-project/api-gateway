import {post, getToken, postFlow} from './common.js';

const CLIENT_ID = '5ffee54b-f82f-4328-a055-ff75fcc1fe85';
const CLIENT_SECRET = 'su_D~HHsCAE8-UUb_pOE6XpYRQ';
const SCOPES = 'read';
const AUTH_URL = 'https://oauth2.ks-kyma-806s.goatz.shoot.canary.k8s-hana.ondemand.com/oauth2/token';
const API_URL = 'https://hello-world-oauth2.ks-kyma-806s.goatz.shoot.canary.k8s-hana.ondemand.com/dispatch';

const REQUEST_BODY = {
  scope: SCOPES,
  grant_type: 'client_credentials',
};

export default function (data) {
    postFlow(API_URL);
}