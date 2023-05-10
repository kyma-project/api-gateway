import http from 'k6/http';
import { check } from 'k6';
//import { Rate } from "k6/metrics";
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

var stringified = JSON.stringify(open('./payload-direct.json'));
//export let errorRate = new Rate("errorRate"); //Give the Same in grafana

export function setup() {
    var url = 'https://oauth2.bc-test.goatz.shoot.canary.k8s-hana.ondemand.com/oauth2/token' //use your URL
    let params = {
        headers: {
            'Authorization' : 'Basic zkIPVnWaSQuCpFoJI1xZmkgMqIucqT52kCMV2cw7XO4.ilPpxAUMsaVEYVw7hLU0-TMZeZohD0Fsyzu2uaPLPvs' //change ur creds
        },
    };

    let form_data = {
        grant_type: 'client_credentials',
        scope: 'write'
    };

    let response = http.post(url, form_data,params);
    let token = response.json().access_token;

    return token;
}

export default function (access_token) {
  var url = 'https://hello-world-jwt-istio.bc-test.goatz.shoot.canary.k8s-hana.ondemand.com'

  var now = Date.now();
  var stringified1 = stringified.replace(/STARTTIME/g, now);
  stringified1 = stringified1.replace(/ACCESS_TOKEN/g, access_token);
  
  var payload = JSON.parse(stringified1);
  var params = {
      headers: {
        'Content-Type': 'application/cloudevents+json',
        'Authorization': `Bearer ${access_token}`,
        'ce-type': 'sf-direct',
        'ce-source': 'k6',
        'ce-eventtypeversion': 'v1',
        'ce-specversion': '1.0',
        'ce-id': 'i513578'
      },
    };
  
    let res = http.post(url, payload, params);
    var status = res.status;
    if( status != 200)
    {
      console.log("Response Code---->" + status)
      console.log("Response Body---->" + res.body)
      //errorRate.add(1)
    }
  
    check(res, { 'SuccessFull calls': (r) => r.status == 200 },);
  
  }

  export function handleSummary(data) {
  return {
    "summary.html": htmlReport(data),
  };
}
