FROM tinygo/tinygo as build
COPY /src/cached-auth-filter /cache-filter
WORKDIR /cached-auth-filter
RUN tinygo build -o cache-filter.wasm -target=wasi ./main.go

FROM envoyproxy/envoy:v1.20.0

COPY --from=build /cache-filter/cache-filter.wasm /cache-filter/cache-filter.wasm
RUN chmod a+rx /cache-filter/cache-filter.wasm
