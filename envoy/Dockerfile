FROM envoyproxy/envoy:v1.20.0

RUN chmod a+rw /lua-json
COPY ../cache-filter.wasm /cache-filter/cache-filter.wasm
RUN chmod a+rx /cache-filter/cache-filter.wasm
