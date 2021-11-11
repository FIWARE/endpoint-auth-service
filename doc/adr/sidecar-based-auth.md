# Add authentication via a sidecar-proxy 

## Status

- proposed

## Context

Notifications from the context broker sometimes(f.e. in the [iShare-context](https://dev.ishareworks.org)) require authentication/authorization. The brokers do 
not know about security concerns and should stay free of it.   

## Decision

The auth-headers will be applied by a [sidecar-proxy](https://www.oreilly.com/library/view/designing-distributed-systems/9781491983638/ch02.html). All
outgoing requests from the broker should be routed through the sidecar-proxy, that is responsible for adding the appropriate headers.

## Consequences

- the broker stays free of security concerns
- development cycles of broker and sidecar don't need to be connected
- development tooling of broker and sidecar can be different
- sidecar can be reused for other components
- sidecar can be added as a "plugin" to the broker deployments