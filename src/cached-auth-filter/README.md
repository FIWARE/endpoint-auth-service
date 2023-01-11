# Cached Auth Filter

The cached-auth-filter is an implementation of the [envoy wasm-filter](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/wasm/v3/wasm.proto) that handles the
calls to the [auth-provider](../../doc/AUTHPROVIDER.md) and caches them on a per-endpoint basis inside envoy's [shared-data](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/data_sharing_between_filters)

## Build

The implementation uses the [Tetratelabs proxy wasm go-sdk](https://github.com/tetratelabs/proxy-wasm-go-sdk). There for it requires 
* [Go](https://go.dev/dl/) >= 1.17
* [Tinygo](https://tinygo.org/)

It can be built either directly, via tinygo:
```shell
tinygo build -o cache-filter.wasm -scheduler=none -target=wasi ./main.go
```
or with a docker-container:
```shell
docker run -v /etc/ssl/certs:/etc/ssl/certs -v $(pwd)/:/cache-filter --workdir /cache-filter tinygo/tinygo:0.26.0 tinygo build -o cache-filter.wasm -scheduler=none -target=wasi ./main.go
```
> :bulb: The tinygo image is very reduced, therefor does not contain the ca's to trust github. They need to be mounted to the container.

## Configuration

The filter supports two working-modes:
* handle everything
* handle configured endpoints

### Handle everything

In this case, the filter will handle all requests forwarded to it. This is mode should either be used if all requests from an application should be handled the same way(e.g. every outgoing request gets an auth-header) or if the filter is used with something like [envoy's matching api](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/matching/matching_api), so that the requests are pre-filtered, before they enter the cached-auth-filter.
Be aware that the MatchingApi currently is an experimental-feature, that needs to be enabled explictly. See for example: [docker-compose setup](docker-compose/initial-config/envoy.yaml). The [endpoint-configuration-service](src/endpoint-configuration-service/) supports configuration-generation for the MatchingApi.


### Handle configured endpoints

When this mode is enabled, the filter will check if the requested endpoint is configured and only handle it in this case. Everything else will be passed-by.

### Configuration reference

```json

    {
        // general plugin configuration
        "general": {
            // timeout to be used when authprovider is requested
            "authRequestTimeout": 5000,
            // address of the authprovider, depending on the proxy it will be a cluster-name(envoy) or something like an upstream
            "authProviderName": "ext-authz",
            // authtype to request
            "authType": "ISHARE",
            // should the filter do endpoint matching
            "enableEndpointMatching" : false
        },
        // configuration for endpoint matching
        "endpoints": {
            // auth type to be used for the following endpoints
            "<AUTHTYPE>-1": {
                // domain to be used with the authtype
                "<DOMAIN-1>": 
                    // paths to be handled. Be aware: if an endpoint is configured twice for multiple auth-types, only the last one will be used.
                    ["/path"],
                "<DOMAIN-2>": ["/path"]
            },
            "<AUTHTYPE>-2": {
                "<DOMAIN-3>": ["/"]
            }
        }
    }

```