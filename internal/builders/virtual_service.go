package builders

import (
	networkingv1alpha1 "knative.dev/pkg/apis/istio/common/v1alpha1"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

// VirtualService returns builder for knative.dev/pkg/apis/istio/v1alpha3/VirtualService type
func VirtualService() *virtualService {
	return &virtualService{
		value: &networkingv1alpha3.VirtualService{},
	}
}

type virtualService struct {
	value *networkingv1alpha3.VirtualService
}

func (vs *virtualService) Get() *networkingv1alpha3.VirtualService {
	return vs.value
}

func (vs *virtualService) From(val *networkingv1alpha3.VirtualService) *virtualService {
	vs.value = val
	return vs
}

func (vs *virtualService) Name(val string) *virtualService {
	vs.value.Name = val
	return vs
}

func (vs *virtualService) GenerateName(val string) *virtualService {
	vs.value.Name = ""
	vs.value.GenerateName = val
	return vs
}

func (vs *virtualService) Namespace(val string) *virtualService {
	vs.value.Namespace = val
	return vs
}

func (vs *virtualService) Owner(val *ownerReference) *virtualService {
	vs.value.OwnerReferences = append(vs.value.OwnerReferences, *val.Get())
	return vs
}

func (vs *virtualService) Label(key, val string) *virtualService {
	if vs.value.Labels == nil {
		vs.value.Labels = make(map[string]string)
	}
	vs.value.Labels[key] = val
	return vs
}

func (vs *virtualService) Spec(val *virtualServiceSpec) *virtualService {
	vs.value.Spec = *val.Get()
	return vs
}

// VirtualServiceSpec returns builder for knative.dev/pkg/apis/istio/v1alpha3/VirtualServiceSpec type
func VirtualServiceSpec() *virtualServiceSpec {
	return &virtualServiceSpec{
		value: &networkingv1alpha3.VirtualServiceSpec{},
	}
}

type virtualServiceSpec struct {
	value *networkingv1alpha3.VirtualServiceSpec
}

func (vss *virtualServiceSpec) Get() *networkingv1alpha3.VirtualServiceSpec {
	return vss.value
}

func (vss *virtualServiceSpec) From(val *networkingv1alpha3.VirtualServiceSpec) *virtualServiceSpec {
	vss.value = val
	return vss
}

func (vss *virtualServiceSpec) Host(val string) *virtualServiceSpec {
	vss.value.Hosts = append(vss.value.Hosts, val)
	return vss
}

func (vss *virtualServiceSpec) Gateway(val string) *virtualServiceSpec {
	vss.value.Gateways = append(vss.value.Gateways, val)
	return vss
}

func (vss *virtualServiceSpec) HTTP(hr *httpRoute) *virtualServiceSpec {
	vss.value.HTTP = append(vss.value.HTTP, *hr.Get())
	return vss
}

// HTTPRoute returns builder for knative.dev/pkg/apis/istio/v1alpha3/HTTPRoute type
func HTTPRoute() *httpRoute {
	return &httpRoute{
		value: &networkingv1alpha3.HTTPRoute{},
	}
}

type httpRoute struct {
	value *networkingv1alpha3.HTTPRoute
}

func (hr *httpRoute) Get() *networkingv1alpha3.HTTPRoute {
	return hr.value
}

func (hr *httpRoute) Match(mr *matchRequest) *httpRoute {
	hr.value.Match = append(hr.value.Match, *mr.Get())
	return hr
}

func (hr *httpRoute) Route(rd *routeDestination) *httpRoute {
	hr.value.Route = append(hr.value.Route, *rd.Get())
	return hr
}

func (hr *httpRoute) CorsPolicy(cc *corsPolicy) *httpRoute {
	hr.value.CorsPolicy = cc.Get()
	return hr
}

// MatchRequest returns builder for knative.dev/pkg/apis/istio/v1alpha3/HTTPMatchRequest type
func MatchRequest() *matchRequest {
	return &matchRequest{
		value: &networkingv1alpha3.HTTPMatchRequest{},
	}
}

type matchRequest struct {
	value *networkingv1alpha3.HTTPMatchRequest
}

func (mr *matchRequest) Get() *networkingv1alpha3.HTTPMatchRequest {
	return mr.value
}

func (mr *matchRequest) URI() *stringMatch {
	mr.value.URI = &networkingv1alpha1.StringMatch{}
	return &stringMatch{mr.value.URI, func() *matchRequest { return mr }}
}

type stringMatch struct {
	value  *networkingv1alpha1.StringMatch
	parent func() *matchRequest
}

func (st *stringMatch) Regex(val string) *matchRequest {
	st.value.Regex = val
	return st.parent()
}

// RouteDestination returns builder for knative.dev/pkg/apis/istio/v1alpha3/HTTPRouteDestination type
func RouteDestination() *routeDestination {
	return &routeDestination{&networkingv1alpha3.HTTPRouteDestination{Weight: 100}}
}

type routeDestination struct {
	value *networkingv1alpha3.HTTPRouteDestination
}

func (rd *routeDestination) Get() *networkingv1alpha3.HTTPRouteDestination {
	return rd.value
}

func (rd *routeDestination) Host(val string) *routeDestination {
	rd.value.Destination.Host = val
	return rd
}

func (rd *routeDestination) Port(val uint32) *routeDestination {
	rd.value.Destination.Port.Number = val
	return rd
}

// CorsPolicy returns builder for knative.dev/pkg/apis/istio/v1alpha3/CorsPolicy type
func CorsPolicy() *corsPolicy {
	return &corsPolicy{
		value: &networkingv1alpha3.CorsPolicy{},
	}
}

type corsPolicy struct {
	value *networkingv1alpha3.CorsPolicy
}

func (cp *corsPolicy) Get() *networkingv1alpha3.CorsPolicy {
	return cp.value
}

func (cp *corsPolicy) AllowHeaders(val ...string) *corsPolicy {
	if len(val) == 0 {
		cp.value.AllowHeaders = nil
	} else {
		cp.value.AllowHeaders = append(cp.value.AllowHeaders, val...)
	}
	return cp
}

func (cp *corsPolicy) AllowMethods(val ...string) *corsPolicy {
	if len(val) == 0 {
		cp.value.AllowMethods = nil
	} else {
		cp.value.AllowMethods = append(cp.value.AllowMethods, val...)
	}
	return cp
}

func (cp *corsPolicy) AllowOrigins(val ...string) *corsPolicy {
	if len(val) == 0 {
		cp.value.AllowOrigin = nil
	} else {
		cp.value.AllowOrigin = append(cp.value.AllowOrigin, val...)
	}
	return cp
}
