kubectl config set-context --current --namespace=load-test
export CLIENT_ID="$(kubectl get secret oauthclient -o jsonpath='{.data.client_id}' | base64 --decode)"
export CLIENT_SECRET="$(kubectl get secret oauthclient -o jsonpath='{.data.client_secret}' | base64 --decode)"
export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
export TOKEN=$(curl -s -X POST "https://oauth2.$DOMAIN/oauth2/token" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "grant_type=client_credentials" -F "scope=read" | jq -r '.access_token')
export K6POD="$(kubectl get pods --selector=app=k6 -o jsonpath='{.items[0].metadata.name}')"

cat > k6_script_auth_gen.js <<EOF
import http from 'k6/http';
import { check } from 'k6';
//import { Rate } from "k6/metrics";
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";

var stringified = JSON.stringify(open('./payload-direct.json'));
//export let errorRate = new Rate("errorRate"); //Give the Same in grafana

export function setup() {
    var url = 'https://oauth2.$DOMAIN/oauth2/token' //use your URL
    let params = {
        headers: {
            'Authorization' : 'Basic $TOKEN' //change ur creds
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
  var url = 'https://hello-world-jwt-istio.$DOMAIN'

  var now = Date.now();
  var stringified1 = stringified.replace(/STARTTIME/g, now);
  stringified1 = stringified1.replace(/ACCESS_TOKEN/g, access_token);
  
  var payload = JSON.parse(stringified1);
  var params = {
      headers: {
        'Content-Type': 'application/cloudevents+json',
        'Authorization': \`Bearer \${access_token}\`,
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
EOF

kubectl cp k6_script_auth_gen.js $K6POD:. -c k6-alpine
kubectl cp payload-direct.json $K6POD:. -c k6-alpine

kubectl exec -it deployment/goat-test-k6 -- k6 run k6_script_auth_gen.js -d 5m --rps 100 --out influxdb=http://goat-test-influxdb:8086/k6 --system-tags=method,name,status,tag
kubectl cp "$K6POD":summary.html summary.html
