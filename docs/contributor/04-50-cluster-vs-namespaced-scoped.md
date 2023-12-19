# Pros and cons of cluster scoped and namespaced scoped module custom resource
In the following overview, we have labelled some items with `[UX]' as they evaluate how the user sees and interacts with the resources. Some of these items are opinions rather than facts as we have no data on them and they are mainly based on our assumptions of how a user should interact with our components and APIs. Also, some of these points are based on our experience with the current implementation of the user interface.
Furthermore, we believe that it is important not only to look at this from a technical perspective, but also to focus on the user experience.

<table style="table-layout: fixed; width: 100%">
    <tbody>
        <tr>
            <td colspan="2" align="center">Cluster scoped Custom Resource</td>
            <td colspan="2" align="center">Namespaced scoped Custom Resource</td>
        </tr>
        <tr>
            <td style="width: 25%;" align="center">Pros</td>
            <td style="width: 25%;" align="center">Cons</td>
            <td style="width: 25%;" align="center">Pros</td>
            <td style="width: 25%;" align="center">Cons</td>
        </tr>
        <tr>
            <td>[UX] Resource is more visible to customer, and the module is clearly visible as installed.<br></br> This might also guarantee better visibility that this resource manages resources across the cluster, not only in a single namespace.</td>
            <td>[UX] Different approach than other Kyma modules</td>
            <td>[UX] Reconciled resource in same namespace as manager</td>
            <td>[UX] Module is less visible in Dashboard, as user needs to access kyma-system namespace, before seeing the resource</td>
        </tr>
        <tr>
            <td>Ability to use cross namespace owner references:<br></br> - Allows getting all managed resources by owner reference<br></br>- Allows easy to do force deletion of the module, even without a controller</td>
            <td>[UX] Module CR ties to Kyma is not visible by namespace, but by CR API Group only.</td>
            <td>[UX] Resource is a clear part of Kyma, since it's in the "kyma-system" namespace.</td>
            <td>Creation and modifying of resources in `kyma-system` is required, even in LM scenario</td>
        </tr>
        <tr>
            <td>[UX] User does not need to create or edit resource in "kyma-system", which should only be managed by Kyma team</td>
            <td>Requires additional configuration to have it displayed in Busola under `kyma-system`, if we consider that to be the desired behaviour</td>
            <td></td>
            <td>Reconciliation for resources in different Namespaces is anyway needed, as they need to be put in `Warning` state</td>
        </tr>
        <tr>
            <td>[UX] Configuring RBAC for creation/visibility of the module is easier (in a cluster that has users with different level of access, e.g. administrator, developer, user)<br></br>This allows a more fine grained approach to configuring ClusterRoles without the need to expose access to `kyma-system`</td>
            <td></td>
            <td></td>
            <td>Owner references don't bring any value, as they would only allow access to `kyma-system` resources</td>
        </tr>
        <tr>
            <td>Monitoring number of resources is easier, and doesn't need to have access to `kyma-system`</td>
            <td></td>
            <td></td>
            <td>Control over Cluster scoped resources by a Namespaced scope resource feels counter-intuitive</td>
        </tr>
        <tr>
            <td>Better solution for enforcing single module installation policy on cluster</td>
            <td></td>
            <td></td>
            <td>[UX] Customer needs to care about the namespace for module CR, when he doesn't want to use the default.</td>
        </tr>
        <tr>
            <td>If the module installation creates other namespaces (for example Istio-system in Istio module) it feels counter intuitive that the managing resource is not higher in NS hierarchy</td>
            <td></td>
            <td></td>
            <td>[UX] Restriction of module CR to `kyma-system` namespace is not obvious for the user. Because it is a namespaced resource, but it is restricted to be reconciled from a specific namespace only.</td>
        </tr>
        <tr>
            <td></td>
            <td></td>
            <td></td>
            <td>Unnecessary fine-grained granularity: The module CR does not only manage resources in the specific namespace, but on the whole cluster.</td>
        </tr>
    </tbody>
</table>
