# Envoy

[Envoy](https://www.envoyproxy.io) is used as the side-car to add auth-information for outgoing requests. 
See the ["Use envoy as proxy"-ADR](adr/choose-proxy.md) for the rational of that choice.

In order to provide an out-of-the-box working proxy, the [FIWARE-specific envoy image](quay.io/fiware/envoy) is provided. It has the wasm-filter already 
built-in.
If another envoy-deployment is used, the wasm-file needs to be provided and its path has to be configured as part of the [endpoint-configuration-service' config](../src/endpoint-configuration-service).
