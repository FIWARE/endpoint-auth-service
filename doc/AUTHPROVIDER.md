# Auth-Provider

An auth-provider is an implementation of the [auth-provider-api](../api/auth-provider-api.yaml) to provide the headers used for auth to the [envoy-proxy](https://www.envoyproxy.io).
It allows support of different auth-mechanisms and its concrete handling independent from the proxy. 

The auth-provider only needs to provide one endpoint: ```GET /{provider}/auth```. The provider parameter allows routing for different auth-types. The provider itself
has to take care about required configuration methods or the concrete handling of secrets. See the [iShare-implementation](../src/ishare-auth-provider) for a concrete example.

