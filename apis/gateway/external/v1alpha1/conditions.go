/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ProcessingConditions returns the full set of conditions that signal reconciliation is in progress.
// Use this at the start of a reconcile loop before any sub-resource work has been done.
func ProcessingConditions(generation int64) []metav1.Condition {
	return []metav1.Condition{
		{
			Type:               ConditionTypeReady,
			Status:             metav1.ConditionUnknown,
			ObservedGeneration: generation,
			Reason:             ReasonReconciling,
			Message:            "Reconciliation in progress",
		},
		{
			Type:               ConditionTypeDNSEntryReady,
			Status:             metav1.ConditionUnknown,
			ObservedGeneration: generation,
			Reason:             ReasonReconciling,
			Message:            "Reconciliation in progress",
		},
		{
			Type:               ConditionTypeCertificateReady,
			Status:             metav1.ConditionUnknown,
			ObservedGeneration: generation,
			Reason:             ReasonReconciling,
			Message:            "Reconciliation in progress",
		},
		{
			Type:               ConditionTypeGatewayConfigured,
			Status:             metav1.ConditionUnknown,
			ObservedGeneration: generation,
			Reason:             ReasonReconciling,
			Message:            "Reconciliation in progress",
		},
	}
}

// WaitingCondition returns the Ready condition used while polling for sub-resource readiness.
func WaitingCondition(generation int64) metav1.Condition {
	return metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionUnknown,
		ObservedGeneration: generation,
		Reason:             ReasonReconciling,
		Message:            "Waiting for sub-resources to become ready",
	}
}

// ReadyCondition returns the Ready condition used when all sub-resources are reconciled successfully.
func ReadyCondition(generation int64) metav1.Condition {
	return metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: generation,
		Reason:             ReasonReady,
		Message:            "All resources reconciled successfully",
	}
}

// ErrorCondition returns the Ready condition used when reconciliation fails.
func ErrorCondition(generation int64, msg string) metav1.Condition {
	return metav1.Condition{
		Type:               ConditionTypeReady,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: generation,
		Reason:             ReasonFailed,
		Message:            msg,
	}
}
