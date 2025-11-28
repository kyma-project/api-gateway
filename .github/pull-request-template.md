<!-- Thank you for your contribution. Before you submit the pull request:
1. Follow contributing guidelines, templates, the recommended Git workflow, and any related documentation.
2. Read and submit the required Contributor Licence Agreements (https://github.com/kyma-project/community/blob/main/docs/contributing/02-contributing.md#agreements-and-licenses).
3. Test your changes and attach their results to the pull request.
4. Update the relevant documentation.
-->

**Description**

Changes proposed in this pull request:

- ...

**Pre-Merge Checklist**

Consider all the following items. If your contribution violates any of them, or you are not sure about it, add a comment to the PR.

- [ ] The code coverage is acceptable.
- [ ] Release notes for the introduced changes are created.
- [ ] If Kubebuilder changes were made, you ran `make manifests` and committed the changes before the merge.
- [ ] Pre-existing managed resources are correctly handled.
- [ ] The change works on all hyperscalers supported by SAP BTP, Kyma runtime.
- [ ] There is no upgrade downtime.
- [ ] For infrastructure changes, you checked if the changes affect the hyperscaler's costs.
- [ ] RBAC settings are as restrictive as possible.
- [ ] If any new libraries are added, you verified license compliance and maintainability, and made a comment in the PR with details. We only allow stable releases to be included in the project.
- [ ] You checked if this change should be cherry-picked to active release branches.
- [ ] The configuration does not introduce any additional latency.
- [ ] If Busola updates are needed.

**Related issues**
<!-- If you refer to a particular issue, provide its number. For example, `Resolves #123`, `Fixes #43`, or `See also #33`. -->
