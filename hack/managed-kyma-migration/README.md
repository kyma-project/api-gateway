# SAP BTP, Kyma runtime migration

> **NOTE**: This documentation is relevant for SAP BTP, Kyma runtime only and does not apply to open-source Kyma.

## Scenarios

### Provisioning of API Gateway CR using Lifecycle Manager in a new cluster

If there is no APIGateway custom resource (CR), then Lifecycle Manager provisions the default API Gateway CR defined in the API Gateway ModuleTemplate. The migration
adds the API Gateway module to the Kyma CR.

### Provisioning of Istio CR using Lifecycle Manager in a cluster with existing modules

If there is no APIGateway CR, then Lifecycle Manager provisions the default APIGateway CR defined in the APIGateway ModuleTemplate. The migration
adds the API Gateway module to the Kyma CR without overwriting existing module configuration.

## Migration test process

### Test scenarios

Apply the ModuleTemplate for both `fast` and `regular` channels to Dev Control Plane.

#### SAP BTP, Kyma runtime clusters without existing modules

1. Create a Dev SAP BTP, Kyma runtime cluster.
2. Execute the migration.
3. Verify that `api-gateway-manager` is installed and the APIGateway CR's status is `Ready`.
4. Verify that the `api-gateway` deployment is not present.

#### SAP BTP, Kyma runtime cluster with an existing module

1. Create a Dev SAP BTP, Kyma runtime cluster.
2. Add the Keda module to the Kyma CR.
   ```yaml
   spec:
     modules:
       - name: keda
   ```
3. Execute the migration.
4. Verify that `api-gateway-manager` is installed and the APIGateway CR's status is `Ready`.
5. Verify that the `api-gateway` deployment is not present.

## Module's rollout and migration

### Preparations

Executing `kcp taskrun` requires the path to the kubeconfig file of the corresponding Gardener project and permissions to list/get shoots.

### Dev

#### Prerequisites

- Reconciliation is disabled for the Dev environment. See PR #4485 in the `kyma/management-plane-config` repository.

#### Migration procedure

1. Apply the ModuleTemplate for both `fast` and `regular` channels to Dev Control Plane.
2. Verify that the ModuleTemplate in the `fast` and `regular` channels is available on SAP BTP, Kyma runtime clusters of the Dev environment.
3. Use `kcp login` to log in to Dev and run the migration script on all SAP BTP, Kyma runtime clusters. To do that, you can use the following command:
   ```shell
   kcp taskrun --gardener-kubeconfig {PATH TO GARDENER PROJECT KUBECONFIG} --gardener-namespace kyma-dev -t all -- ./managed-kyma-migration.sh
   ```
4. Verify that the migration worked as expected by checking the status of APIGateway manifests on Control Plane.
   ```shell
   kubectl get manifests -n kcp-system -o custom-columns=NAME:metadata.name,STATE:status.state | grep api-gateway
   ```
5. Verify that the `api-gateway` deployment is not present on any of the SAP BTP, Kyma runtime clusters.
   ```shell
   kcp taskrun --gardener-kubeconfig {PATH TO GARDENER PROJECT KUBECONFIG} --gardener-namespace kyma-dev -t all -- kubectl get deployment -n kyma-system api-gateway 2>/dev/null
   ```

### Stage

Perform the rollout to Stage together with the SRE team. Since they have already performed the rollout for other modules, they might suggest a different rollout strategy.

#### Prerequisites

- Reconciliation is disabled for the Stage environment. See PR #4486 to the `kyma/management-plane-config` repository.

#### Migration procedure

1. Apply the ModuleTemplate for both `fast` and `regular` channels to Stage Control Plane.
2. Verify that the ModuleTemplate in the `fast` and `regular` channels is available in SAP BTP, Kyma runtime clusters of the Stage environment.
3. Use `kcp login` to log in to Stage, select a few SAP BTP, Kyma runtime clusters on `Kyma-Test/Kyma-Integration`, and run `managed-kyma-migration.sh` on them using `kcp taskrun`.
4. Verify if the migration was successful on the SAP BTP, Kyma runtime clusters by checking the status of APIGateway CR and the reconciler's components.
5. Run `managed-kyma-migration.sh` for all SKRs in `Kyma-Test` and `Kyma-Integration` global accounts.
6. Verify if the migration worked as expected.
7. Run `managed-kyma-migration.sh` for the whole Canary landscape.
8. Verify if the migration worked as expected.

### Prod

Perform the rollout to Prod together with the SRE team. Since they have already performed the rollout for other modules, they might suggest a different rollout strategy.

#### Prerequisites

- Reconciliation is disabled for the Prod environment. See PR #4487 to the `kyma/management-plane-config` repository.

#### Migration procedure

1. Commit the module manifest to the `regular` and `fast` channels in the `kyma/module-manifests` internal repository.
2. Verify that the ModuleTemplates are present in the `kyma/kyma-modules` internal repository.
3. Verify that the ModuleTemplate in both channels are available on `Prod` environment SKRs.
4. Use `kcp login` to log in to Prod, select some SAP BTP, Kyma runtime clusters on `Kyma-Test/Kyma-Integration`, and run `managed-kyma-migration.sh` on them using `kcp taskrun`.
5. Verify if the migration was successful on the SAP BTP, Kyma runtime clusters by checking the status of APIGateway CR and the reconciler's components.
6. Run `managed-kyma-migration.sh` for all SAP BTP, Kyma runtime clusters in Trial global accounts.
7. Verify if the migration worked as expected.
8. Run `managed-kyma-migration.sh` for the whole Factory landscape.
9. Verify if the migration worked as expected.
