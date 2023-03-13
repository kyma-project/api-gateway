import http from 'k6/http';
import { check, group } from 'k6';
import encoding from 'k6/encoding';
import { Trend,Counter } from 'k6/metrics';

let getRequestTrend = new Trend('get_request_duration', true);
let postRequestTrend = new Trend('post_request_duration', true);
let refreshTokenCount = new Counter('refresh-tokens');
let newTokenCounter = new Counter('initial-token');
let tokenData = {};

export function postFlow(url) {
    group('get-initial-token', function () {
        tokenData = getNewOrRefreshToken(tokenData);
    });
    group('post-request', function () {
        post(API_URL, tokenData);
    })
}

export function get(url, tokenData) {
    let params = {
        headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${tokenData.auth.access_token}`,
        },
    };

    let resp = http.get(url, params);
    getRequestTrend.add(resp.timings.duration);
    if (!check(resp, {
        'is status 200': (r) => r.status === 200,
    })){
        console.log(`status: ${resp.status}`);
        console.log(JSON.stringify(resp));
    }

    //check refresh token 
    getNewOrRefreshToken(tokenData);
}

export function post(url, tokenData) {
    const payload = JSON.stringify({
        'x': 'y'
    });
    const params = {
        headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${tokenData.auth.access_token}`,
        },
    };
    let resp = http.post(url, payload, params);
    postRequestTrend.add(resp.timings.duration);
    if (!check(resp, {
        'is status 200': (r) => r.status === 200,
    })){
        console.log(`status: ${resp.status}`);
        console.log(JSON.stringify(resp));
    }

    //check refresh token 
    getNewOrRefreshToken(tokenData);
}

export function getNewOrRefreshToken(tokenData) {
    var t = new Date();
    t.setSeconds(t.getSeconds() + 30);

    if(tokenData.auth === undefined){
        tokenData.auth = getToken(CLIENT_ID, CLIENT_SECRET, REQUEST_BODY, AUTH_URL, false);
    } else if (t.getTime() >= tokenData.auth.expiresAt) {
        tokenData.auth = getToken(CLIENT_ID, CLIENT_SECRET, REQUEST_BODY, AUTH_URL, true);
    }

    return tokenData;
}

export function getToken(clientId, clientSecret, requestBody, url, isTokenRefresh) {

    var t = new Date();

    const encodedCredentials = encoding.b64encode(
        `${clientId}:${clientSecret}`,
    );

    const params = {
        headers: {
            Authorization: `Basic ${encodedCredentials}`,
        },
    };

    const resp = http.post(url, requestBody, params);
    if (isTokenRefresh) {
        refreshTokenCount.add(1);
    } else {
        newTokenCounter.add(1);
    }
    if(!check(resp, {
        'successfully got token': (r) => r.status === 200,
    })){
        console.log(`status: ${resp.status}`);
        console.log(JSON.stringify(resp));
    }

    let response = resp.json();
    response.expiresAt = t.setSeconds(t.getSeconds() + response.expires_in);
    return response;
}