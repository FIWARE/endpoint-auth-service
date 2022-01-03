# Implement the auth mechanism as a wasm-filter

## Status

- proposed

## Context

Auth handling requires caching of auth-information to be performant, instead of running through the whole flow each request.

## Decision

The auth-handling will be implemented as a [wasm-filter](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/wasm/v3/wasm.proto) and use 
[envoy's shared data feature](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/data_sharing_between_filters) for caching the auth information.

## Rational

- shared-data is a built-in feature from envoy, that does not require additional cache-functionallity
- wasm-filters are "normal" http-filters that are natively supported by envoy
- wasm-filters can be written in multiple languages
- through the [matching api](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/matching/matching_api), the filters can be applied at a route level