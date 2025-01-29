# Module Status Handling

## Status
Proposed

## Context

Currently, APIGateway module sets the state `Processing` on the APIGateway Custom Resource whenever there is reconciliation / installation happening.
From the monitoring and kubernetes API standpoint, this is not easy to properly handle,
as there is no possibility to observe module readiness, signified by module CR being in `Ready` state,
without having in mind the periodic switch to the `Processing` state.

## Proposal

This ADR proposes changes to handling the state of the module CR, with three different solutions possible:

### Solution 1

Proposal: Remove the `Processing` state entirely from the module API.

Consequences: From technical standpoint, this would be a breaking change, as there might be users relying on the `Processing` state,
for example, to determine if the module is being installed or not.
However, as the purpose of the `Processing` state is not clear, it might be a good idea to remove it entirely.

### Solution 2

Proposal: Do a `soft` removal of the `Processing` state. 
This would mean that this state is still present in the API, but will never be set by the module.

Consequences: This would be a less breaking change than the first solution, as the `Processing` state would still be present in the API,
but the installation would not be possible to observe in this state.

### Solution 3

Proposal: Improve on the state transition logic.
A simple solution would be not to switch the module state to `Processing` after `Ready` or `Error` has occurred.
This means that the `Processing` state would only be set initially when installing the module for the first time.

Consequences: As the initial installation could be considered the most important moment to observe the `Processing` state,
this solution would be a good compromise between the removal of the state and keeping it in the API.
However, this would mean that the `Processing` state handling logic cannot be entirely removed from the module.

### Solution 4

Proposal: Same as 3, but switch back to `Processing` when user changes configuration of the APIGateway Custom Resource.

Consequences: This is technically harder to implement as it requires retention of the previously applied state.


## Decision

The team decided to start with solution 3 with possibility to extend the logic further to solution 4 later, if needed.
This decision was discussed with lead Kyma Architect, and as a general guidance, the processing state should only happen in case the module might not be available (a possible downtime), which means it is NOT ready. As is the case for APIGateway, this state should generally never occur unless the module user changes the configuration (e.g. disabling default Kyma gateway), so the module should be considered `Ready` almost always.
