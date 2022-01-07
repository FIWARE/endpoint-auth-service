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
docker run -v $(pwd)/:/cache-filter --workdir /cache-filter tinygo/tinygo tinygo build -o cache-filter.wasm -scheduler=none -target=wasi ./main.go
```
## Configuration

