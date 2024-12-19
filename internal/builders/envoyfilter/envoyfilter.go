package envoyfilter

import (
	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	apiv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

type ConfigPatch = networkingv1alpha3.EnvoyFilter_EnvoyConfigObjectPatch

type Builder struct {
	name          string
	namespace     string
	selectors     map[string]string
	annotations   map[string]string
	ConfigPatches []*ConfigPatch
}

// WithName sets name of EnvoyFilter
func (e *Builder) WithName(name string) *Builder {
	e.name = name
	return e
}

// WithNamespace sets namespace of EnvoyFilter
func (e *Builder) WithNamespace(namespace string) *Builder {
	e.namespace = namespace
	return e
}

// WithAnnotation adds annotation to the EnvoyFilter. Can be used multiple times to add more annotations
func (e *Builder) WithAnnotation(key, val string) *Builder {
	if e.annotations == nil {
		e.annotations = make(map[string]string)
	}
	e.annotations[key] = val
	return e
}

// WithWorkloadSelector adds labels to EnvoyFilter's WorkloadSelector. Can be used multiple times to add more selectors
func (e *Builder) WithWorkloadSelector(key, val string) *Builder {
	if e.selectors == nil {
		e.selectors = make(map[string]string)
	}
	e.selectors[key] = val
	return e
}

// WithConfigPatch adds provided patch to the end of the EnvoyFilter patches chain
func (e *Builder) WithConfigPatch(patch *ConfigPatch) *Builder {
	e.ConfigPatches = append(e.ConfigPatches, patch)
	return e
}

// Build returns EnvoyFilter generated from the configuration provided to the builder.
func (e *Builder) Build() *apiv1alpha3.EnvoyFilter {
	f := apiv1alpha3.EnvoyFilter{}
	if len(e.name) > 0 {
		f.Name = e.name
	}
	if len(e.namespace) > 0 {
		f.Namespace = e.namespace
	}
	if len(e.selectors) > 0 {
		f.Spec.WorkloadSelector = &networkingv1alpha3.WorkloadSelector{
			Labels: e.selectors,
		}
	}

	if len(e.annotations) > 0 {
		f.Annotations = e.annotations
	}

	f.Spec.ConfigPatches = e.ConfigPatches
	return &f
}

// NewEnvoyFilterBuilder returns EnvoyFilterBuilder for building istio EnvoyFilters
func NewEnvoyFilterBuilder() *Builder {
	return &Builder{}
}
