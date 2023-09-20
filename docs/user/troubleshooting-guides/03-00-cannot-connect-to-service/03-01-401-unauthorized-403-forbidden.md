---
title: 401 Unauthorized or 403 Forbidden
---

## Symptom

When you try to reach your Service, you get `401 Unauthorized` or `403 Forbidden` in response.

## Remedy

Make sure that the following conditions are met:

- You are using an access token with proper scopes, and it is active:

  1. Export the credentials of your OAuth2Client as environment variables:

      > **NOTE:** Export the **CLIENT_NAMESPACE** and **CLIENT_NAME** variables before you proceed with step 1.
      
      ```bash
      export CLIENT_ID="$(kubectl get secret -n $CLIENT_NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_id}' | base64 --decode)"
      export CLIENT_SECRET="$(kubectl get secret -n $CLIENT_NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_secret}' | base64 --decode)"
      ```

  2. Encode your client credentials and export them as an environment variable:

      ```bash
      export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
      ```

  3. Check the access token status:

      ```bash
      curl -X POST "https://oauth2.{CLUSTER_DOMAIN}/oauth2/introspect" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "token={ACCESS_TOKEN}"
      ```

  4. Generate a [new access token](../../tutorials/01-50-expose-and-secure-a-workload/01-50-expose-and-secure-workload-oauth2.md) if needed.