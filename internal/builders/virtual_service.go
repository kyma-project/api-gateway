package builders

import (
	apirulev1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	"google.golang.org/protobuf/types/known/durationpb"
	"istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"strconv"
	"strings"
	"time"
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

func (hr *httpRoute) Headers(h *v1beta1.Headers) *httpRoute {
	hr.value.Headers = h
	return hr
}

func (hr *httpRoute) Timeout(value time.Duration) *httpRoute {
	hr.value.Timeout = durationpb.New(value)
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

func (st *stringMatch) Prefix(val string) *matchRequest {
	st.value.MatchType = &v1beta1.StringMatch_Prefix{Prefix: val}
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

// NewHttpRouteHeadersBuilder returns builder for istio.io/api/networking/v1beta1/Headers type
func NewHttpRouteHeadersBuilder() HttpRouteHeadersBuilder {
	return HttpRouteHeadersBuilder{
		value: &v1beta1.Headers{
			Request: &v1beta1.Headers_HeaderOperations{
				Set: make(map[string]string),
			},
		},
	}
}

type HttpRouteHeadersBuilder struct {
	value *v1beta1.Headers
}

func (h HttpRouteHeadersBuilder) Get() *v1beta1.Headers {
	return h.value
}

func (h HttpRouteHeadersBuilder) SetHostHeader(hostname string) HttpRouteHeadersBuilder {
	h.value.Request.Set["x-forwarded-host"] = hostname
	return h
}

// SetRequestCookies sets the Cookie header and expects a string of the form "cookie-name1=cookie-value1; cookie-name2=cookie-value2; ..."
func (h HttpRouteHeadersBuilder) SetRequestCookies(cookies string) HttpRouteHeadersBuilder {
	h.value.Request.Set["Cookie"] = cookies
	return h
}

// SetRequestHeaders sets the request headers and expects a map of the form "header-name1": "header-value1", "header-name2": "header-value2", ...
func (h HttpRouteHeadersBuilder) SetRequestHeaders(headers map[string]string) HttpRouteHeadersBuilder {
	for name, value := range headers {
		h.value.Request.Set[name] = value
	}

	return h
}

const (
	ExposeName       = "Access-Control-Expose-Headers"
	AllowHeadersName = "Access-Control-Allow-Headers"
	CredentialsName  = "Access-Control-Allow-Credentials"
	AllowMethodsName = "Access-Control-Allow-Methods"
	OriginName       = "Access-Control-Allow-Origin"
	MaxAgeName       = "Access-Control-Max-Age"
)

// SetHeader sets the request header with name and value
func (h HttpRouteHeadersBuilder) SetCORSPolicyHeaders(corsPolicy apirulev1beta1.CorsPolicy) HttpRouteHeadersBuilder {
	if len(corsPolicy.ExposeHeaders) > 0 {
		h.value.Response.Set[ExposeName] = strings.Join(corsPolicy.ExposeHeaders, ",")
	} else {
		h.value.Response.Remove = append(h.value.Request.Remove, ExposeName)
	}

	if len(corsPolicy.AllowHeaders) > 0 {
		h.value.Response.Set[AllowHeadersName] = strings.Join(corsPolicy.AllowHeaders, ",")
	} else {
		h.value.Response.Remove = append(h.value.Request.Remove, AllowHeadersName)
	}

	if corsPolicy.AllowCredentials != nil {
		if *corsPolicy.AllowCredentials {
			h.value.Response.Set[CredentialsName] = "true"
		} else {
			h.value.Response.Set[CredentialsName] = "false"
		}
	} else {
		h.value.Response.Remove = append(h.value.Request.Remove, CredentialsName)
	}

	if len(corsPolicy.AllowMethods) > 0 {
		h.value.Response.Set[AllowMethodsName] = strings.Join(corsPolicy.AllowMethods, ",")
	} else {
		h.value.Response.Remove = append(h.value.Request.Remove, AllowMethodsName)
	}

	if len(corsPolicy.AllowOrigins) > 0 {
		h.value.Response.Set[OriginName] = strings.Join(corsPolicy.AllowOrigins, ",")
	} else {
		h.value.Response.Remove = append(h.value.Request.Remove, OriginName)
	}

	if corsPolicy.MaxAge != nil {
		h.value.Response.Set[MaxAgeName] = strconv.Itoa(int(corsPolicy.MaxAge.Seconds()))
	} else {
		h.value.Response.Remove = append(h.value.Request.Remove, MaxAgeName)
	}

	return h
}
