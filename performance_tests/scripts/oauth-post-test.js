import {post, getToken, postFlow} from './common.js';

const CLIENT_ID = `${ CLIENT_ID }`;
const CLIENT_SECRET = `${ CLIENT_SECRET }`;
const SCOPES = 'read';
const AUTH_URL = `${ AUTH_URL }`;
const API_URL = `${ API_URL }`;

const REQUEST_BODY = {
  scope: SCOPES,
  grant_type: 'client_credentials',
};

export default function (data) {
    postFlow(API_URL);
}
