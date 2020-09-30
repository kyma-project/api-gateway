package builders

import (
	"istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

// VirtualService returns builder for istio.io/client-go/pkg/apis/networking/v1beta1/VirtualService type
func VirtualService() *virtualService {
	return &virtualService{
		value: &networkingv1beta1.VirtualService{},
	}
}

type virtualService struct {
	value *networkingv1beta1.VirtualService
}

func (vs *virtualService) Get() *networkingv1beta1.VirtualService {
	return vs.value
}

func (vs *virtualService) From(val *networkingv1beta1.VirtualService) *virtualService {
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

// VirtualServiceSpec returns builder for istio.io/api/networking/v1beta1/VirtualServiceSpec type
func VirtualServiceSpec() *virtualServiceSpec {
	return &virtualServiceSpec{
		value: &v1beta1.VirtualService{},
	}
}

type virtualServiceSpec struct {
	value *v1beta1.VirtualService
}

func (vss *virtualServiceSpec) Get() *v1beta1.VirtualService {
	return vss.value
}

func (vss *virtualServiceSpec) From(val *v1beta1.VirtualService) *virtualServiceSpec {
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
	vss.value.Http = append(vss.value.Http, hr.Get())
	return vss
}

// HTTPRoute returns builder for istio.io/api/networking/v1beta1/HTTPRoute type
func HTTPRoute() *httpRoute {
	return &httpRoute{
		value: &v1beta1.HTTPRoute{},
	}
}

type httpRoute struct {
	value *v1beta1.HTTPRoute
}

func (hr *httpRoute) Get() *v1beta1.HTTPRoute {
	return hr.value
}

func (hr *httpRoute) Match(mr *matchRequest) *httpRoute {
	hr.value.Match = append(hr.value.Match, mr.Get())
	return hr
}

func (hr *httpRoute) Route(rd *routeDestination) *httpRoute {
	hr.value.Route = append(hr.value.Route, rd.Get())
	return hr
}

func (hr *httpRoute) CorsPolicy(cc *corsPolicy) *httpRoute {
	hr.value.CorsPolicy = cc.Get()
	return hr
}

// MatchRequest returns builder for istio.io/api/networking/v1beta1/HTTPMatchRequest type
func MatchRequest() *matchRequest {
	return &matchRequest{
		value: &v1beta1.HTTPMatchRequest{},
	}
}

type matchRequest struct {
	value *v1beta1.HTTPMatchRequest
}

func (mr *matchRequest) Get() *v1beta1.HTTPMatchRequest {
	return mr.value
}

func (mr *matchRequest) Uri() *stringMatch {
	mr.value.Uri = &v1beta1.StringMatch{}
	return &stringMatch{mr.value.Uri, func() *matchRequest { return mr }}
}

type stringMatch struct {
	value  *v1beta1.StringMatch
	parent func() *matchRequest
}

func (st *stringMatch) Regex(val string) *matchRequest {
	st.value.MatchType = &v1beta1.StringMatch_Regex{Regex: val}
	return st.parent()
}

// RouteDestination returns builder for istio.io/api/networking/v1beta1/HTTPRouteDestination type
func RouteDestination() *routeDestination {
	return &routeDestination{&v1beta1.HTTPRouteDestination{
		Destination: &v1beta1.Destination{
			Port: &v1beta1.PortSelector{},
		},
		Weight: 100,
	}}
}

type routeDestination struct {
	value *v1beta1.HTTPRouteDestination
}

func (rd *routeDestination) Get() *v1beta1.HTTPRouteDestination {
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

// CorsPolicy returns builder for istio.io/api/networking/v1beta1/CorsPolicy type
func CorsPolicy() *corsPolicy {
	return &corsPolicy{
		value: &v1beta1.CorsPolicy{},
	}
}

type corsPolicy struct {
	value *v1beta1.CorsPolicy
}

func (cp *corsPolicy) Get() *v1beta1.CorsPolicy {
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

func (cp *corsPolicy) AllowOrigins(val ...*v1beta1.StringMatch) *corsPolicy {
	if len(val) == 0 {
		cp.value.AllowOrigins = nil
	} else {
		cp.value.AllowOrigins = append(cp.value.AllowOrigins, val...)
	}
	return cp
}
