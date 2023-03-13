# Support for Ory Oathkeeper mutators with Istio

## Templating support

Ory rules have support for templating in mutators. Simple cases can be implemented with the help of envoy filter, such as the one presented in [id-token-envoyfilter](./id-token-envoyfilter) directory.

## Mutators

### Headers

`Headers` type mutator can be handled with Istio with the help of Virtual Service [HeaderOperations](https://istio.io/latest/docs/reference/config/networking/virtual-service/#Headers-HeaderOperations). With HeaderOperations it is only possible to add static data, but in the case of [Ory Headers Mutator](https://www.ory.sh/docs/oathkeeper/pipeline/mutator#headers) templating is supported which receives the current Authentication Session. To support similar capabilities in Istio an [EnvoyFilter](https://istio.io/latest/docs/reference/config/networking/envoy-filter/) must be used.

Ory configuration:

```yaml
...
mutators:
  - config:
      headers:
        X-Some-Arbitrary-Data: "test"
    handler: header
...
```

Coresponding Istio Virtual Service configuration:

```yaml
spec:
  http:
    - headers:
        request:
          set:
            X-Some-Arbitrary-Data: "test"
```

### Cookie

The mutator of type `Cookie` can be handled the same as the mutator [Headers](#headers) with Istio using Virtual Service HeaderOperations. Here applies the same limitiations for Istio as only static data can be added, but [Ory Cookie Mutator](https://www.ory.sh/docs/oathkeeper/pipeline/mutator#cookie) supports templating.

Ory configuration:

```yaml
...
mutators:
  - config:
      cookies:
        user: "test"
    handler: cookie
...
```

Coresponding Istio Virtual Service configuration:

```yaml
spec:
  http:
    - headers:
        request:
          set:
            Cookie: "user=test"
```

### Id_token

It seems not to be possible to support the same functionality as `id_token` mutator as this requires a mechanism for encoding and signing the response from OAuth2 server (e.g. Ory Hydra) into a JWT. 
The other issue is that the JWKS used for signing this JWT is deployed as a secret `ory-oathkeeper-jwks-secret` that would have to be fetched in context of the implementation or mounted into the component doing the encoding.

### Hydrator

Support for Hydrator token would require to call external APIs in context of Istio proxy. This mutator also influences other mutators by running before others and supplying them with the outcome of running.
