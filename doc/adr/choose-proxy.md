# Use envoy as proxy

## Status

- proposed

## Context

We need a proxy to handle outgoing requests. The usual runtime environment will be cloud-based, most probably with kubernetes as a cluster-manager.

## Decision

The proxy to be used will be [envoy](https://www.envoyproxy.io/).

## Consequences

- is maintained by a wide community, most notably the [CNCF](https://www.cncf.io/),  [Lyft](https://www.lyft.com/) and [Google](https://www.google.com/).
- is build for cloud purposes
- very fast according to various benchmarks, thus limiting the impact on the broker performance
- provides a lua-based scripting mechanism to add custom headers