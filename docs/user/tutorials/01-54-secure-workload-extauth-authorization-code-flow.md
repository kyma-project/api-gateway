# Expose and Secure a Workload with exAuth Using the Authorization Code Flow

## Prerequisites

- You have an SAP BTP, Kyma runtime instance with the Istio and API Gateway modules added. The Istio and API Gateway modules are added to your Kyma cluster by default.
- You have an SAP Cloud Identity Services tenant. See [Initial Setup](https://help.sap.com/docs/cloud-identity-services/cloud-identity-services/initial-setup?locale=en-US&version=Cloud&q=open+id+connect).

## Procedure


### Create and Configure OpenID Connect Application

You need an identity provider to issue JWTs. Creating an OpenID Connect application allows SAP Cloud Identity Services to act as your issuer and manage authentication for your workloads.

1. Sign in to the administration console for SAP Cloud Identity Services. See [Access Admin Console](https://help.sap.com/docs/cloud-identity-services/cloud-identity-services/accessing-administration-console?locale=en-US&version=Cloud).

2. Create an OpenID Connect Application.

   1. Go to **Application Resources** > **Application**.
   2. Choose **Create**, provide the application name, and select the OpenID Connect radio button. 
      For more configuration options, see [Create OpenID Connect Application](https://help.sap.com/docs/cloud-identity-services/cloud-identity-services/create-openid-connect-application-299ae2f07a6646768cbc881c4d368dac?locale=en-US&version=Cloud).
   3. Choose **+Create**.

3. Configure OpenID Connect Application for the Authorization Code flow.
   
   1. In the **Trust > Single Sign-On** section of your created application, choose **OpenID Connect Configuration**.
   2. Provide the name.
   3. Add the `https://oauth2-proxy.{kyma-domain)/oauth2/callback` redirect URI. //TODO 
   3. In the **Grant types** section, check **Authorization Code**.
      For more configuration options, see [Configure OpenID Connect Application for Authorization Code Flow]().
   4. Choose **Save**.

4. Configure a secret for API authentication.

   1. In the **Application API** section of your created application, choose **Client Authentication**.
   2. In the **Secrets** section, choose **Add**.
   3. Choose the OpenID API access and provide other configuration as needed.
      For more configuration options, see [Configure Secrets for API Authentication](https://help.sap.com/docs/cloud-identity-services/cloud-identity-services/dev-configure-secrets-for-api-authentication?version=Cloud&locale=en-US).
   4. Choose **Save**.
      Your client ID and secret appear in a pop-up window. Save the secret, as you will not be able to retrieve it from the system later.