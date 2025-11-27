# Tutorials - Expose and Secure a Workload
Browse the API Gateway tutorials to learn how to expose and secure workloads.

> [!WARNING]
> APIRule CRD `v2` is the latest stable version. Version `v1beta1` is removed in release 3.4 of the API Gateway module. 
>- All existing `v1beta1` APIRule configurations continue to function as expected.
>- APIRules `v1beta1` are no longer visible in the Kyma dashboard. You can still display them using kubectl, but the resources are displayed in the converted `v2` format.
>- You can no longer create new `v1beta1` APIRules, delete or edit the existing ones. To make any changes, you must migrate to version `v2`.
>
> **Required action**: Migrate all your APIRule custom resources (CRs) to version `v2`. For the detailed migration procedure, see [APIRule Migration](../../apirule-migration/README.md). For the complete deletion timeline for SAP BTP, Kyma runtime, see [APIRule Migration Timeline](https://help.sap.com/docs/btp/sap-business-technology-platform/apirule-migration?locale=en-US#apirule-v1beta1-migration-timeline).

Browse the API Gateway tutorials to learn how to expose workloads.

- [Get a JSON Web Token (JWT)](./01-51-get-jwt.md)
- [Expose and Secure a Workload with JWT](./01-52-expose-and-secure-workload-jwt.md)
- [Expose and Secure a Workload with extAuth](./01-53-expose-and-secure-workload-ext-auth.md)
- [Use the XFF Header to Configure IP-Based Access to a Workload](./01-55-ip-based-access-with-xff.md)
