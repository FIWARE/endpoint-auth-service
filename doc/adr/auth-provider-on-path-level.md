# Auth-provider config on path level

## Status

- proposed

## Context

The sidecar-proxy should be able to support multiple auth-methods. 

## Decision

The target auth-provider for each request will be configured on path-level, rather then on full-url or domain-level.

## Rational

- configuring on a path level allows envoy to reuse the connections, thus making it more performant
- configuration on a domain level would allow the configurer to insert any (potentially unsafe) auth-provider, 
  without a proper option to check from the system's operator -> potential vulnerabiltiy
- if the auth-providers reside under different domains, the central "ext-authz" cluster can point to a load balancer that does
  the actual routing, while it only servers configured providers that the operator "knows" -> same flexibility, less potential for misuse