# Get a JSON Web Token (JWT)

This tutorial shows how to get a JSON Web Token (JWT), which can be used to access secured endpoints created in the [Expose and secure a workload with Istio](./01-53-expose-and-secure-workload-istio.md) and [Expose and secure a workload with JWT](./01-52-expose-and-secure-workload-jwt.md) tutorials.

## Prerequisites

You use an OpenID Connect-compliant (OIDC-compliant) identity provider.

## Steps

1. Create an application in your OIDC-compliant identity provider. Save the client credentials: Client ID and Client Secret. 

2. In the URL `https://{YOUR_IDENTITY_PROVIDER_INSTANCE}/.well-known/openid-configuration` replace `{YOUR_IDENTITY_PROVIDER_INSTANCE}` with the name of your OICD-compliant identity provider instance. Then, open the link in your browser. Save the values of the **token_endpoint**, **jwks_uri** and **issuer** parameters.

3. To encode your client credentials and get a JWT, use [Postman](https://www.postman.com) or [curl](https://curl.se).

<!-- tabs:start -->
#### **Postman**

1. Enter your token endpoint URL.
2. Go to the `Body` tab and select the `x-www-form-urlencoded` option. Add two key-value pairs to the body:
   - **grant_type**: `client_credentials`
   - **client_id**: `{CLIENT_ID}`
   
   Replace `{CLIENT_ID}` with your client ID.
2. Go to the `Headers` tab and add the header:
   - **Content-Type**: `application/x-www-form-urlencoded`
3. Go to the `Authorization` tab and select:
   - **Type**: Basic
   - **Username**: `{CLIENT_ID}`
   - **Password**: `{CLIENT_SECRET}`
   
   Replace `{CLIENT_ID}` and `{CLIENT_SECRET}` with your Client ID and Client Secret.
4. Send a `POST` request. 

#### **curl**

1. Export the saved values as environment variables:
   
   ```bash
   export CLIENT_ID={YOUR_CLIENT_ID}
   export CLIENT_SECRET={YOUR_CLIENT_SECRET}
   export TOKEN_ENDPOINT={YOUR_TOKEN_ENDPOINT}
   ```

2. Encode your client credentials and export them as an environment variable:

   ```bash
   export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
   ```

3. To obtain a JWT, run the following command:

   ```bash
   curl -X POST "$TOKEN_ENDPOINT" -d "grant_type=client_credentials" -d "client_id=$CLIENT_ID" -H "Content-Type: application/x-www-form-urlencoded" -H "Authorization: Basic $ENCODED_CREDENTIALS"
   ```
<!-- tabs:end -->

4. Save your JWT. 

To learn how to secure a workload using a JWT, follow [Expose and Secure a Workload with JWT](./01-52-expose-and-secure-workload-jwt.md) and 
[Expose and Secure a Workload with Istio](./01-53-expose-and-secure-workload-istio.md).