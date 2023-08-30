//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIGateway) DeepCopyInto(out *APIGateway) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIGateway.
func (in *APIGateway) DeepCopy() *APIGateway {
	if in == nil {
		return nil
	}
	out := new(APIGateway)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *APIGateway) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIGatewayList) DeepCopyInto(out *APIGatewayList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]APIGateway, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIGatewayList.
func (in *APIGatewayList) DeepCopy() *APIGatewayList {
	if in == nil {
		return nil
	}
	out := new(APIGatewayList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *APIGatewayList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIGatewaySpec) DeepCopyInto(out *APIGatewaySpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIGatewaySpec.
func (in *APIGatewaySpec) DeepCopy() *APIGatewaySpec {
	if in == nil {
		return nil
	}
	out := new(APIGatewaySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIGatewayStatus) DeepCopyInto(out *APIGatewayStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIGatewayStatus.
func (in *APIGatewayStatus) DeepCopy() *APIGatewayStatus {
	if in == nil {
		return nil
	}
	out := new(APIGatewayStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIRule) DeepCopyInto(out *APIRule) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIRule.
func (in *APIRule) DeepCopy() *APIRule {
	if in == nil {
		return nil
	}
	out := new(APIRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *APIRule) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIRuleList) DeepCopyInto(out *APIRuleList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]APIRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIRuleList.
func (in *APIRuleList) DeepCopy() *APIRuleList {
	if in == nil {
		return nil
	}
	out := new(APIRuleList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *APIRuleList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIRuleResourceStatus) DeepCopyInto(out *APIRuleResourceStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIRuleResourceStatus.
func (in *APIRuleResourceStatus) DeepCopy() *APIRuleResourceStatus {
	if in == nil {
		return nil
	}
	out := new(APIRuleResourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIRuleSpec) DeepCopyInto(out *APIRuleSpec) {
	*out = *in
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(Service)
		(*in).DeepCopyInto(*out)
	}
	if in.Gateway != nil {
		in, out := &in.Gateway, &out.Gateway
		*out = new(string)
		**out = **in
	}
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]Rule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIRuleSpec.
func (in *APIRuleSpec) DeepCopy() *APIRuleSpec {
	if in == nil {
		return nil
	}
	out := new(APIRuleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIRuleStatus) DeepCopyInto(out *APIRuleStatus) {
	*out = *in
	if in.LastProcessedTime != nil {
		in, out := &in.LastProcessedTime, &out.LastProcessedTime
		*out = (*in).DeepCopy()
	}
	if in.APIRuleStatus != nil {
		in, out := &in.APIRuleStatus, &out.APIRuleStatus
		*out = new(APIRuleResourceStatus)
		**out = **in
	}
	if in.VirtualServiceStatus != nil {
		in, out := &in.VirtualServiceStatus, &out.VirtualServiceStatus
		*out = new(APIRuleResourceStatus)
		**out = **in
	}
	if in.AccessRuleStatus != nil {
		in, out := &in.AccessRuleStatus, &out.AccessRuleStatus
		*out = new(APIRuleResourceStatus)
		**out = **in
	}
	if in.RequestAuthenticationStatus != nil {
		in, out := &in.RequestAuthenticationStatus, &out.RequestAuthenticationStatus
		*out = new(APIRuleResourceStatus)
		**out = **in
	}
	if in.AuthorizationPolicyStatus != nil {
		in, out := &in.AuthorizationPolicyStatus, &out.AuthorizationPolicyStatus
		*out = new(APIRuleResourceStatus)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIRuleStatus.
func (in *APIRuleStatus) DeepCopy() *APIRuleStatus {
	if in == nil {
		return nil
	}
	out := new(APIRuleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Authenticator) DeepCopyInto(out *Authenticator) {
	*out = *in
	if in.Handler != nil {
		in, out := &in.Handler, &out.Handler
		*out = new(Handler)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Authenticator.
func (in *Authenticator) DeepCopy() *Authenticator {
	if in == nil {
		return nil
	}
	out := new(Authenticator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Handler) DeepCopyInto(out *Handler) {
	*out = *in
	if in.Config != nil {
		in, out := &in.Config, &out.Config
		*out = new(runtime.RawExtension)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Handler.
func (in *Handler) DeepCopy() *Handler {
	if in == nil {
		return nil
	}
	out := new(Handler)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Mutator) DeepCopyInto(out *Mutator) {
	*out = *in
	if in.Handler != nil {
		in, out := &in.Handler, &out.Handler
		*out = new(Handler)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Mutator.
func (in *Mutator) DeepCopy() *Mutator {
	if in == nil {
		return nil
	}
	out := new(Mutator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Rule) DeepCopyInto(out *Rule) {
	*out = *in
	if in.Methods != nil {
		in, out := &in.Methods, &out.Methods
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AccessStrategies != nil {
		in, out := &in.AccessStrategies, &out.AccessStrategies
		*out = make([]*Authenticator, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Authenticator)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Mutators != nil {
		in, out := &in.Mutators, &out.Mutators
		*out = make([]*Mutator, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Mutator)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Rule.
func (in *Rule) DeepCopy() *Rule {
	if in == nil {
		return nil
	}
	out := new(Rule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Service) DeepCopyInto(out *Service) {
	*out = *in
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(uint32)
		**out = **in
	}
	if in.Host != nil {
		in, out := &in.Host, &out.Host
		*out = new(string)
		**out = **in
	}
	if in.IsExternal != nil {
		in, out := &in.IsExternal, &out.IsExternal
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Service.
func (in *Service) DeepCopy() *Service {
	if in == nil {
		return nil
	}
	out := new(Service)
	in.DeepCopyInto(out)
	return out
}
