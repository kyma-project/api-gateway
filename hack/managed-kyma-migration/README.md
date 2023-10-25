# Managed Kyma migration

> **NOTE**: This documentation is relevant for managed Kyma only and does not apply to OS Kyma.

## Scenarios

### Provisioning of API Gateway CR via Lifecycle-Manager in a new cluster

If there is no APIGateway CR, then Lifecycle Manager provisions the default API Gateway CR defined in the API Gateway ModuleTemplate. The migration
adds the API Gateway module to the Kyma CR.

### Provisioning of Istio CR via Lifecycle Manager in a cluster with existing modules

If there is no APIGateway CR, then Lifecycle Manager provisions the default APIGateway CR defined in the APIGateway ModuleTemplate. The migration
adds the API Gateway module to the Kyma CR without overwriting existing module configuration.

## Migration test process

### Test scenarios

1. Apply the ModuleTemplate for both `fast` and `regular` channel to DEV Control Plane.

#### SKR without existing modules

1. Create Dev SKR.
2. Execute the migration.
3. Verify that `api-gateway-manager` is installed and the APIGateway CR's status is `Ready`.
4. Verify that `api-gateway` deployment is not present.

#### SKR with an existing module

1. Create Dev SKR.
2. Add the Keda module to the Kyma CR.
   ```yaml
   spec:
     modules:
       - name: keda
   ```
3. Execute the migration.
4. Verify that `api-gateway-manager` is installed and the APIGateway CR's status is `Ready`.
5. Verify that `api-gateway` deployment is not present.

## Module's rollout and migration

### Preparations

1. Executing `kcp taskrun` requires the path to the kubeconfig file of the corresponding Gardener project and permissions to list/get shoots.

### Dev

1. Apply the module template to Dev Control Plane.
2. `kcp login` to Dev and run the migration script on all SKRs. To do that, you can use the following command:
   ```shell
   kcp taskrun --gardener-kubeconfig {PATH TO GARDENER PROJECT KUBECONFIG} --gardener-namespace kyma-dev -t all -- ./managed-kyma-migration.sh
   ```
3. Verify that the migration worked as expected by checking the status of APIGateway manifests on Control Plane.
   ```shell
   kubectl get manifests -n kcp-system -o custom-columns=NAME:metadata.name,STATE:status.state | grep api-gateway
   ```
4. Verify that `api-gateway` deployment is not present on any of the SKRs
   ```shell
   kcp taskrun --gardener-kubeconfig {PATH TO GARDENER PROJECT KUBECONFIG} --gardener-namespace kyma-dev -t all -- kubectl get deployment -n kyma-system api-gateway 2>/dev/null
   ```

### Stage

Perform the rollout to Stage together with the SRE team. Since they have already performed the rollout for other modules, they might suggest a different rollout strategy.

#### Prerequisites

- Reconciliation is disabled for Stage environment

#### Migration procedure

1. Commit the module manifest to the `kyma/module-manifests` internal repository, using the `fast` channel.
2. Verify that the ModuleTemplate is present in `kyma/kyma-modules` internal repository.
3. Verify that the ModuleTemplate in `fast` channel is available on `Stage` environment SKRs.
4. `kcp login` to Stage, select some SKRs on `Kyma-Test/Kyma-Integration`, and run `managed-kyma-migration.sh` on them using `kcp taskrun`.
5. Verify if the migration was successful on the SKRs by checking the status of APIGateway CR and the reconciler's components.
6. Run `managed-kyma-migration.sh` for all SKRs in Kyma-Test and Kyma-Integration global accounts.
7. Verify if the migration worked as expected.
8. Run `managed-kyma-migration.sh` for the whole Canary landscape.
9. Verify if the migration worked as expected.

### Prod

Perform the rollout to `Prod` together with the SRE team. Since they have already performed the rollout for other modules, they might suggest a different rollout strategy.

#### Prerequisites

- Reconciliation is disabled for Prod environment

#### Migration procedure

1. Promote the module manifest prior commited to the `fast` channel in the `kyma/module-manifests` internal repository to the `regular` channel.
2. Verify that the ModuleTemplate is present in `kyma/kyma-modules` internal repository.
3. Verify that the ModuleTemplate in `regular` channel is available on `Prod` environment SKRs.
4. `kcp login` to Prod, select some SKRs on `Kyma-Test/Kyma-Integration`, and run `managed-kyma-migration.sh` on them using `kcp taskrun`.
5. Verify if the migration was successful on the SKRs by checking the status of APIGateway CR and the reconciler's components.
6. Run `managed-kyma-migration.sh` for all SKRs in Trial global accounts.
7. Verify if the migration worked as expected.
8. Run `managed-kyma-migration.sh` for the whole Factory landscape.
9. Verify if the migration worked as expected.
