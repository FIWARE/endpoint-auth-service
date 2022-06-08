FROM tinygo/tinygo as build
COPY /src/cached-auth-filter /cached-auth-filter
WORKDIR /cached-auth-filter
RUN tinygo build -o cached-auth-filter.wasm -target=wasi ./main.go

FROM envoyproxy/envoy:v1.21.2

COPY --from=build /cached-auth-filter/cached-auth-filter.wasm /cached-auth-filter/cached-auth-filter.wasm
RUN chmod a+rx /cached-auth-filter/cached-auth-filter.wasm
